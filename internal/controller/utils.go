package controller

import (
	"context"
	"fmt"
	"reflect"

	"encoding/json"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetUnstructuredObject returns an unstructured object given its name and namespace
func GetUnstructuredObject(c client.Client, name, namespace string) (*unstructured.Unstructured, error) {
	unstructuredObj := &unstructured.Unstructured{}
	objKey := client.ObjectKey{Name: name, Namespace: namespace}
	if err := c.Get(context.Background(), objKey, unstructuredObj); err != nil {
		return nil, err
	}
	return unstructuredObj, nil
}

// ListUnstructuredObjectsByLabels returns a list of unstructured objects with given type
// and with given labels to search for
func ListUnstructuredObjectsByLabels(c client.Client, searchLabels map[string]string, refType map[string]string) (*unstructured.UnstructuredList, error) {
	// Iterate over the list of objects with given group, version and kind
	// and search for the object with the given labels
	unstructuredObjList := &unstructured.UnstructuredList{}
	unstructuredObjList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   refType["group"],
		Version: refType["version"],
		Kind:    refType["kind"],
	})
	if err := c.List(context.Background(), unstructuredObjList, client.MatchingLabels(searchLabels)); err != nil {
		return nil, err
	}
	return unstructuredObjList, nil
}

// ListUnstructuredObjectsByFieldList returns a list of unstructured objects with given type
// and with given field path and its value to search for
func ListUnstructuredObjectsByFieldList(
	c client.Client, searchSpec map[string]string, refType map[string]string, fields ...string) (*unstructured.UnstructuredList, error) {
	// Iterate over the list of objects with given group, version and kind
	// and search for the object with the given specs within the dependedBy list
	filteredObj := &unstructured.UnstructuredList{}
	unstructuredObjList := &unstructured.UnstructuredList{}
	unstructuredObjList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   refType["group"],
		Version: refType["version"],
		Kind:    refType["kind"],
	})
	if err := c.List(context.Background(), unstructuredObjList); err != nil {
		return nil, err
	}
	for _, obj := range unstructuredObjList.Items {
		if objSpec, err := GetNestedField(obj.Object, fields[:len(fields)-1]...); err != nil {
			return nil, err
		} else {
			descriptors := objSpec[fields[len(fields)-1]]
			switch d := descriptors.(type) {
			case []any:
				for _, descriptor := range d {
					if descriptorMap, err := ObjectToStringMap(descriptor); err != nil {
						return nil, err
					} else {
						if CompareStringMap(descriptorMap, searchSpec) {
							filteredObj.Items = append(filteredObj.Items, obj)
						}
					}
				}
			default:
				return nil, errors.New(fmt.Sprintf("field %s is not a list", fields[len(fields)-1]))
			}
		}
	}
	return filteredObj, nil
}

// GetUnstructuredConditionByType retrieves the condition with the given type from the unstructured object
func GetUnstructuredConditionByType(obj *unstructured.Unstructured, t string) (bool, map[string]any, error) {
	conditions, err := GetNestedValue(obj.Object, "status", "conditions")
	if err != nil {
		return false, nil, err
	}
	if _, ok := conditions.([]any); !ok {
		return false, nil, fmt.Errorf("conditions is not of type []any")
	}
	cds := conditions.([]any)
	// Find the Ready condition in each of the objects
	cdIdx := IndexOfMapValue(cds, "type", t)
	if cdIdx == -1 {
		return false, nil, nil
	}
	cd := cds[cdIdx].(map[string]any)
	return true, cd, nil
}

// ParseConditionStatus returns the metav1.ConditionStatus based on the given status string
func ParseConditionStatus(status any) metav1.ConditionStatus {
	if status == nil {
		return metav1.ConditionUnknown
	}
	switch status {
	case "True":
		return metav1.ConditionTrue
	case "False":
		return metav1.ConditionFalse
	default:
		return metav1.ConditionUnknown
	}
}

// IndexOfTypedCondition finds the index of the given key in the list of conditions
func IndexOfTypedCondition(list []metav1.Condition, key string) int {
	for i, item := range list {
		if item.Type == key {
			return i
		}
	}
	return -1
}

// GetTypedCondition retrieves the condition with the given type from the list of conditions
func GetTypedCondition(conditions []metav1.Condition, t string) (bool, *metav1.Condition) {
	// Find the Ready condition in each of the objects
	cdIdx := IndexOfTypedCondition(conditions, t)
	if cdIdx == -1 {
		return false, nil
	}
	cd := conditions[cdIdx]
	return true, &cd
}

// GetTypedConditionStatus retrieves the status of condition with the given type from the list of conditions
// returns the condition value if it exists and nil otherwise
func GetTypedConditionStatus(conditions []metav1.Condition, t string) *metav1.ConditionStatus {
	// Find the Ready condition in each of the objects
	cdIdx := IndexOfTypedCondition(conditions, t)
	if cdIdx == -1 {
		return nil
	}
	cd := conditions[cdIdx]
	return &cd.Status
}

// SetTypedCondition sets the condition with the given type in the list of conditions
func SetTypedCondition(conditions []metav1.Condition, t string, status metav1.ConditionStatus, reason, message string, tt metav1.Time) []metav1.Condition {
	// Find the Ready condition in each of the objects
	cdIdx := IndexOfTypedCondition(conditions, t)
	if cdIdx == -1 {
		conditions = append(conditions, metav1.Condition{
			LastTransitionTime: tt,
			Type:               t,
			Status:             status,
			Reason:             reason,
			Message:            message,
		})
	} else {
		conditions[cdIdx].LastTransitionTime = tt
		conditions[cdIdx].Status = status
		conditions[cdIdx].Reason = reason
		conditions[cdIdx].Message = message
	}
	return conditions
}

// RemoveFromTypedCondition removes the condition with the given type from the list
func RemoveFromTypedCondition(list []metav1.Condition, key string) []metav1.Condition {
	for i, item := range list {
		if item.Type == key {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}

// GetNestedField retrieves a nested map within a map[string]any structure.
// It traverses the object using the provided sequence of field names.
// Example:
//
//	nested, err := GetNestedField(obj, "spec", "image")
//
// the obj["spec"]["image"] should be a map[string]any otherwise it will return an errors
func GetNestedField(obj map[string]any, fields ...string) (map[string]any, error) {
	if len(fields) == 0 {
		return nil, errors.New("no fields provided")
	}
	m := obj
	for _, field := range fields {
		if val, ok := m[field].(map[string]any); ok {
			m = val
		} else {
			return nil, errors.New(fmt.Sprintf("field [%s] not found in the object or its type is not map[string]any", field))
		}
	}
	return m, nil // the last field is not found in the object
}

// GetNestedValue returns the nested value of a map[string]interface{} object as an interface{}
func GetNestedValue(obj map[string]any, fields ...string) (any, error) {
	f := fields[:len(fields)-1]
	value, err := GetNestedField(obj, f...)
	if err != nil {
		return nil, err
	}
	if val, ok := value[fields[len(fields)-1]]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("field %s not found in the object", fields[len(fields)-1])
}

// HasNestedMap checks if a nested slice of maps within a data structure contains a specified target map.
// It navigates through the `fields` within the `obj` map and looks for `target` in the final slice.
// Returns a boolean indicating existence, the index of the found map, and an error if any occurs.
func HasNestedMap(obj map[string]any, target map[string]string, fields ...string) (bool, int, error) {
	foundIdx, exists := -1, false
	m, err := GetNestedField(obj, fields[:len(fields)-1]...)
	if err != nil {
		return false, foundIdx, err
	}
	field := fields[len(fields)-1]
	switch m[field].(type) {
	case []any:
		valList := m[field].([]any)
		for idx, val := range valList {
			if mapString, err := ObjectToStringMap(val); err != nil {
				return false, foundIdx, err
			} else {
				if CompareStringMap(mapString, target) {
					exists = true
					foundIdx = idx
					break
				}
			}
		}
		return exists, foundIdx, nil
	case nil:
		return false, foundIdx, nil
	default:
		return false, foundIdx, errors.New(fmt.Sprintf("the field %s is not a list", field))
	}
}

// SetNestedField sets the value of a nested field in a map[string]any
// if the latest field is nil, it will be set to the value
// if the latest field is not nil, it will be overwritten
func SetNestedField(obj map[string]any, value any, fields ...string) error {
	m := obj
	for _, field := range fields[:len(fields)-1] {
		if val, ok := m[field]; ok {
			if valMap, ok := val.(map[string]any); ok {
				m = valMap
			} else {
				return errors.New(fmt.Sprintf("field %s is not a map[string]any", field))
			}
		} else {
			newVal := make(map[string]any)
			m[field] = newVal
			m = newVal
		}
	}
	field := fields[len(fields)-1]
	m[field] = value
	return nil
}

// UpdateNestedValue updates the value of a nested field in a map[string]any
// if the latest field is not string, it will return an error
// It returns a boolean indicating if the value was updated and an error if any occurs
func UpdateNestedValue(obj map[string]any, value string, fields ...string) (bool, error) {
	m, err := GetNestedField(obj, fields[:len(fields)-1]...)
	if err != nil {
		return false, err
	}
	field := fields[len(fields)-1]
	if _, ok := m[field].(string); !ok {
		return false, errors.New(fmt.Sprintf("field %s not found in the object or its not a string", field))
	}
	if m[field] == value {
		return false, nil
	}
	m[field] = value
	return true, nil
}

// AppendToNestedList appends a value to the list in the nested field within a map[string]any.
// If the field is nil, it initializes it with a list containing the value.
func AppendToNestedList(obj map[string]any, value any, fields ...string) error {
	m, err := GetNestedField(obj, fields[:len(fields)-1]...)
	if err != nil {
		return err
	}
	field := fields[len(fields)-1]
	switch m[field].(type) {
	case []any:
		m[field] = append(m[field].([]any), value)
	case nil:
		m[field] = []any{value}
	default:
		return errors.New(fmt.Sprintf("field %s not found in the object or its not either nil or a list", field))
	}
	return nil
}

// HasAllLabels checks if all labels are present in the object
func HasAllLabels(objLabels map[string]string, labelKeys []string) bool {
	for _, key := range labelKeys {
		if _, exists := objLabels[key]; !exists {
			return false
		}
	}
	return true
}

// HasAllLabelsAndValue checks if all labels are present in the objLabels and have the same value
func HasAllLabelsAndValue(objLabels map[string]string, labels map[string]string) bool {
	for key, value := range labels {
		if value2, exists := objLabels[key]; !exists || value2 != value {
			return false
		}
	}
	return true
}

// UpdateLabelsIfDifferent updates the objLabels if they are different
// if the objLabels is nil, it will be set to the labels
func UpdateLabelsIfDifferent(objLabels map[string]string, labels map[string]string) {
	if objLabels == nil {
		objLabels = labels
	}
	for key, value := range labels {
		vv, exists := (objLabels)[key]
		if !exists || vv != value {
			objLabels[key] = value
		}
	}
}

// CompareStringSlices returns true if two slices are equal
func CompareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CompareStringMap returns true if two given map[string]string have the same labels and same values
func CompareStringMap(objLabels map[string]string, labels map[string]string) bool {
	keys1 := make([]string, 0, len(objLabels))
	keys2 := make([]string, 0, len(labels))
	if !CompareStringSlices(keys1, keys2) {
		return false
	}
	for key, value := range labels {
		if objLabels[key] != value {
			return false
		}
	}
	return true
}

// StringInSlice returns true if the given string is in the list
func StringInSlice(s string, list []string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// MergeStringMaps merges two maps and returns the result
func MergeStringMaps(a, b map[string]string) map[string]string {
	res := make(map[string]string)
	for k, v := range a {
		res[k] = v
	}
	for k, v := range b {
		res[k] = v
	}
	return res
}

// StructToStringMap converts the given object to a map[string]string
func StructToStringMap(obj any) map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i).Interface()
		result[field.Name] = fmt.Sprintf("%v", value)
	}
	return result
}

// GetConfigMapsByLabels returns a list of ConfigMaps in the given namespace with the given seachLabels
func GetConfigMapsByLabels(c client.Client, namespace string, searchLabels map[string]string) (*corev1.ConfigMapList, error) {
	cmList := &corev1.ConfigMapList{}
	listOptions := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(searchLabels),
	}
	if err := c.List(context.Background(), cmList, listOptions); err != nil {
		return nil, err
	}
	return cmList, nil
}

// GetConfigMap returns a ConfigMap in the given namespace with the given name
func GetConfigMap(c client.Client, name, namespace string) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	key := client.ObjectKey{Name: name, Namespace: namespace}
	if err := c.Get(context.Background(), key, cm); err != nil {
		return nil, err
	}
	return cm, nil
}

// ObjectToMap returns the map[string]any of an object
func ObjectToMap(obj any) (map[string]any, error) {
	fieldBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	// Unmarshal JSON into a map
	var fieldMap map[string]any
	if err := json.Unmarshal(fieldBytes, &fieldMap); err != nil {
		return nil, err
	}
	return fieldMap, nil
}

// ObjectToStringMap returns the map[string]any of an object
func ObjectToStringMap(obj any) (map[string]string, error) {
	fieldBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	// Unmarshal JSON into a map
	var fieldMap map[string]string
	if err := json.Unmarshal(fieldBytes, &fieldMap); err != nil {
		return nil, err
	}
	return fieldMap, nil
}

// ConvertInterfaceMapToStringMap returns a map[string]string from a map[string]interface{}
func ConvertInterfaceMapToStringMap(m map[string]any) map[string]string {
	res := map[string]string{}
	for k, v := range m {
		res[k] = fmt.Sprintf("%v", v)
	}
	return res
}

// SafeString returns the message string based on the given message
func SafeString(data any) string {
	if data == nil {
		return ""
	}
	return data.(string)
}

// RemoveStringAt removes the element at the given index from the list
func RemoveStringAt(list []string, idx int) []string {
	return append(list[:idx], list[idx+1:]...)
}

// IndexOfMapKey finds the index of the given key in the list of maps
func IndexOfMapKey(list []map[string]string, key string) int {
	for i, item := range list {
		if _, ok := item[key]; ok {
			return i
		}
	}
	return -1
}

// IndexOfMapValue finds the index of the given key-value pair in the list of interfaces
// The given key-value pair should be convertible to map[string]string
func IndexOfMapValue(list []any, key string, value string) int {
	for i, item := range list {
		if val, ok := item.(map[string]any)[key]; ok && val.(string) == value {
			return i
		}
	}
	return -1
}

// RemoveNestedListItem removes the element at the given index from the nested list field
func RemoveNestedListItem(obj map[string]any, idx int, fields ...string) error {
	m, err := GetNestedField(obj, fields[:len(fields)-1]...)
	if err != nil {
		return err
	}
	field := fields[len(fields)-1]
	switch m[field].(type) {
	case []any:
		valList := m[field].([]any)
		m[field] = append(valList[:idx], valList[idx+1:]...)
	default:
		return errors.New(fmt.Sprintf("field %s not found in the object or its not a list", field))
	}
	return nil
}
