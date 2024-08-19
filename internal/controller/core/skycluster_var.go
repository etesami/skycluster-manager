package core

var (
	SkyClusterAPI string = "skycluster-manager.savitestbed.ca"

	// managed-by annotation
	SkyClusterAnnotationManagedBy string = SkyClusterAPI + "/managed-by"

	// config-type annotation
	SkyClusterAnnotationConfigType       string = SkyClusterAPI + "/config-type"
	SkyClusterAnnotationProvierName      string = SkyClusterAPI + "/provider-name"
	SkyClusterAnnotationProvierRegion    string = SkyClusterAPI + "/provider-region"
	SkyClusterAnnotationSkyClusterRegion string = SkyClusterAPI + "/skycluster-region"
	SkyClusterAnnotationProvierZone      string = SkyClusterAPI + "/provider-zone"
	SkyClusterAnnotationProvierType      string = SkyClusterAPI + "/provider-type"
	SkyClusterAnnotationCreationTime     string = SkyClusterAPI + "/creation-time"
	SkyClusterAnnotationCompletionTime   string = SkyClusterAPI + "/completion-time"
)
