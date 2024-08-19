package core

var pythonBaseScript string = `
import numpy as np
import collections
import pulp
import json

class Task:
	def __init__(self, name):
		self.name = name
		self.vservices = []
		self.locData = []
		
	def add_vservice(self, vs):
		self.vservices.append(vs)
	
	def get_vservices(self):
		return self.vservices

	def add_locdata(self, loc):
		# loc is a dic {locName: aws-east1, locType: edge, locRegion: west1}
		self.locData.append(loc)

	def get_locdata(self):
		return self.locData

	def contain_vservice(self, vs):
		return vs in self.vservices

class Dag:
	def __init__(self):
		self.tasks = []
		self.vservices = {}
		self.name = None
		self.edge_latencies = {}
		import networkx as nx 
		self.graph = nx.DiGraph()

	def add(self, task):
		self.graph.add_node(task)
		self.tasks.append(task)

	def remove(self, task):
		self.tasks.remove(task)
		self.graph.remove_node(task)

	def add_edge(self, op1, op2, latency=1000):
		# default acceotable latency threshold is set to 1 second, 
		# meaning it is okay to place op1 and op2 anywhere with 
		# latency below 1 second
		assert op1 in self.graph.nodes
		assert op2 in self.graph.nodes
		self.graph.add_edge(op1, op2)
		if not op1.name in self.edge_latencies:
			self.edge_latencies[op1.name] = {}
		self.edge_latencies[op1.name][op2.name] = latency

	def get_graph(self):
		return self.graph

	def get_tasks(self):
		return self.tasks

	def get_edges(self):
		return self.graph.edges

class VService:
	def __init__(self, name, costs=None):
		self.name = name
		self.costs = {}
		if costs is not None:
			for p in costs:
				self.costs[p] = costs[p]
						
	def set_costs(self, costs):
		for p in costs:
			self.costs[p] = costs[p]
					
	def get_costs(self):
		return self.costs

class Provider:
	def __init__(self, name):
		self.name = name
		self.name_base = ""
		self.region = ""
		self.zone = ""
		self.providerType = ""

	def set_provider_name(self, name):
		self.name_base = name
			
	def set_region(self, region):
		self.region = region
	
	def set_zone(self, zone):
		self.zone = zone

	def set_providerType(self, providerType):
		self.providerType = providerType

def convert_to_ms(time_str):
	# Dictionary to define conversion factors to milliseconds
	conversion_factors = {
		'ms': 1,
		's': 1000,
		'ns': 1e-6,
		'us': 1e-3,
	}
	
	# Extract numeric value and unit from the input string
	import re
	match = re.match(r'(\d+(?:\.\d+)?)(\D+)', time_str)
	if not match:
		# raise ValueError("Invalid time format")
		return 0
	
	value, unit = match.groups()
	value = float(value)
	
	# Convert to milliseconds
	if unit in conversion_factors:
		return value * conversion_factors[unit]
	else:
		# raise ValueError("Unsupported time unit")
		return 0
`

var pythonOptimizationBase string = `
prob = pulp.LpProblem('cost_optimization', pulp.LpMinimize)
				
# Prepare the constants.
V = dag.get_tasks()
E = dag.graph.edges()	
`

var pythonOptimizationProblem string = `
# Define the decision variables.
c = {}
for v in dag.get_tasks():
	c[v.name] = pulp.LpVariable.dicts(v.name, providers.keys(), cat='Binary') 

objective = 0
for v in V:
	# [frontend, backend, ...]
	for ppName in c[v.name].keys():
		# optimization variables
		# ppName = vv.name.replace(v.name + '_',"").replace('_','-')
		providerSelected = True
		# print('	 ', ppName, end=' ')
		# virtual services for [frontend, backend, ...]
		task_vservices = v.get_vservices() # [vs1, vs2]
		# we need to sum the cost of virtual services requested
		# and offered by a provider, if a provider offer all services
		# we calculate the sum of the cost for each service it offers
		totalPPCost = 0
		# print(v.name, vv, ppName)
		for tvs in task_vservices:
			# [vs1, vs2]
			# for all the VSs the current provider (ppName)
			# should offer them all, otherwise we discard the provider
			# tvs.get_costs() {provider1: cost, provider2: cost}
			if ppName in tvs.get_costs():
				# if the current optimization variable for provider x
				# offers the virtual service, get its costs
				# print(tvs.name, "is offered in", ppName, "cost: ", tvs.get_costs()[ppName])
				totalPPCost += float(tvs.get_costs()[ppName])
			else:
				providerSelected = False
				# print(ppName, 'does not offer', tvs.name, '[SKIP]')
				# this provider does not offer at least one vs, we skip it
				break
		if providerSelected:
			# The provider offers all VSs and the total cost of 
			# using it is totalPPCost
			# we can write the objective as:
			# print('	 ', c[v.name][ppName], '*', ppName, 'with cost:', totalPPCost)
			objective += c[v.name][ppName] * totalPPCost

# Formulate the constraints.
# c[v] constraints
# the maximum number of instances for each task is the number of providers
# at least one instance of each task is deployed
for v in V:
	prob += pulp.lpSum(c[v.name]) <= len(providers)
	prob += pulp.lpSum(c[v.name]) >= 1

# Linearization constraints
# construct e[u][v][ij]:
# e[u][v][ij] = 1 means task u is assigned to provider i and task v is assigned to provider j
pp_sorted = sorted(list(providers.keys()))
e = collections.defaultdict(lambda: collections.defaultdict(lambda: collections.defaultdict(dict)))
for u, v in E:
	for ii in range(0, len(pp_sorted)-1):
		for jj in range(ii+1, len(pp_sorted)):
			e[u.name][v.name][pp_sorted[ii]+'_'+pp_sorted[jj]] = pulp.LpVariable(
				u.name + v.name + pp_sorted[ii]+'_'+pp_sorted[jj] , cat='Binary')

# 2. e[u][v] 
for u, v in E:
	for ii in range(0, len(pp_sorted)-1):
		for jj in range(ii+1, len(pp_sorted)):
			prob += e[u.name][v.name][pp_sorted[ii]+'_'+pp_sorted[jj]] <= c[u.name][pp_sorted[ii]]
			prob += e[u.name][v.name][pp_sorted[ii]+'_'+pp_sorted[jj]] <= c[v.name][pp_sorted[jj]]
			prob += e[u.name][v.name][pp_sorted[ii]+'_'+pp_sorted[jj]] >= c[v.name][pp_sorted[ii]] + c[v.name][pp_sorted[jj]] - 1
			prob += e[u.name][v.name][pp_sorted[ii]+'_'+pp_sorted[jj]] >= 0
			prob += (
					e[u.name][v.name][pp_sorted[ii]+'_'+pp_sorted[jj]] * 
					(
							provider_latencies[pp_sorted[ii]][pp_sorted[jj]]
					) <= dag.edge_latencies[u.name][v.name]
			)


# Construct F
# C'ij=Cij+Cji
for u, v in E:
	for ii in range(0, len(pp_sorted)-1):
		for jj in range(ii+1, len(pp_sorted)):
			objective += (
						e[u.name][v.name][pp_sorted[ii] + '_' + pp_sorted[jj]] * 
						(
								float(egress_cost_dict[pp_sorted[ii]][pp_sorted[jj]]) + 
								float(egress_cost_dict[pp_sorted[jj]][pp_sorted[ii]])
						)
				)

prob += objective
`

var pythonOptimizationConstraints string = `
# Construct special constraints based on the input data
# Special constraints
# if composite constraints do not meet, then we don't place the task 

for v in V:
	# [frontend, backend, ...]
	# print(v.name, v.get_locdata())
	for pp in providers:
		for ll in v.get_locdata():
			locName = ll['locName']
			locType = ll['locType']
			locRegion = ll['locRegion']
			# print("	 ", pp, end = ' ')
			if ((locName != "" and locName != providers[pp].name_base) or
				(locType != "" and locType != providers[pp].providerType) or
				(locRegion != "" and locRegion != providers[pp].region)):
					# make it zerp
				prob += c[v.name][pp] == 0
`

var pythonOptimizationCommand string = `
__PYTHON_BASE_SCRIPT__

providers = {}
filename = '/providers/__FILE_NAME_PROVIDERS__'
with open(filename, 'r') as file:
	for line in file:
		if line.strip() == "":
			continue
		provider_name = line.strip().split(",")[0]
		pp = Provider(provider_name)
		pp.set_provider_name(line.strip().split(",")[1])
		pp.set_region(line.strip().split(",")[2])
		pp.set_zone(line.strip().split(",")[3])
		pp.set_providerType(line.strip().split(",")[4])
		providers[provider_name] = pp

vservices = {}
vs_costs = {}
filename = '/vservices/__FILE_NAME_VSERVICES__'
with open(filename, 'r') as file:
	for line in file:
		vservice_name = line.strip().split(":")[0]
		vservice_provider = line.strip().split(":")[1].split(",")[0]
		vservice_cost = line.strip().split(":")[1].split(",")[1]
		# print(vservice_name, vservice_provider, vservice_cost)
		if vservice_name not in vservices:
			vservices[vservice_name] = VService(vservice_name)
		vservices[vservice_name].set_costs({vservice_provider: vservice_cost})

filename = '/tasks/__FILE_NAME_TASKS__'
dag = Dag()
tasks = {}
with open(filename, 'r') as file:
	for line in file:
		tt = line.strip().split(":")[0]
		task = Task(tt)
		tasks[tt] = task
		tt_vservices = line.strip().split(":")[1].split("__")[0].split(",")
		for vs in tt_vservices:
			task.add_vservice(vservices[vs])
		tt_locData = line.strip().split(":")[1].split("__")[1:]
		for locData in tt_locData:
			locData = locData.split(",")
			ll = {}
			ll['locName'] = locData[0]
			ll['locType'] = locData[1]
			ll['locRegion'] = locData[2]
			tasks[tt].add_locdata(ll)
		dag.add(task)
		# print(line.strip())

filename = '/edges/__FILE_NAME_EDGES__'
with open(filename, 'r') as file:
	for line in file:
		u = line.strip().split(":")[0]
		v = line.strip().split(":")[1].split(',')[0]
		latency = line.strip().split(":")[1].split(',')[1]
		dag.add_edge(tasks[u],tasks[v], convert_to_ms(latency))

egress_cost_dict = {}
provider_latencies = {}
for pp in providers:
	egress_cost_dict[pp] = {}
	provider_latencies[pp] = {}
	for dd in providers:
		provider_latencies[pp][dd] = ""
		if pp == dd: 
			provider_latencies[pp][dd] = 0
			egress_cost_dict[pp][dd] = 0

filename = '/providerattr/__FILE_NAME_PROVIDERATTR__'
with open(filename, 'r') as file:
	for line in file:
		if line.strip() == "":
			continue
		pp = line.strip().split(':')[0]
		dd = line.strip().split(':')[1].split(',')[0]
		latency = line.strip().split(':')[1].split(',')[1]
		cc = line.strip().split(':')[1].split(',')[2]
		egress_cost_dict[pp][dd] = cc
		provider_latencies[pp][dd] = convert_to_ms(latency)

k = {}
for vs in vservices.values():
		k[vs.name] = vs.get_costs()

__PYTHON_OPTIMIZATION_BASE__

__PYTHON_OPTIMIZATION_PROBLEM__

__PYTHON_OPTIMIZATION_CONSTRAINTS__

# Last Step: Solve the problem
solver = pulp.PULP_CBC_CMD(msg=0)
prob.solve(solver)

# Step 8: Print the results
result = {}
result['status'] = pulp.LpStatus[prob.status]
result['tasks'] = {}

if pulp.LpStatus[prob.status] == 'Optimal':
	for v in V:
		result['tasks'][v.name] = []
		for ll in c[v.name]:
			if c[v.name][ll].varValue == 1:
				res = {}
				res['provider'] = providers[ll].name_base
				res['region'] = providers[ll].region
				res['providerType'] = providers[ll].providerType
				result['tasks'][v.name].append(res)

print(json.dumps(result))

`
