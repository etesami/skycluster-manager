import numpy as np
import collections
import pulp
class Task:
	def __init__(self, name):
		self.name = name
		self.vservices = []

	def add_vservice(self, vs):
		self.vservices.append(vs)
	
	def get_vservices(self):
		return self.vservices

	def contain_vservice(self, vs):
		return vs in self.vservices

class Dag:
	def __init__(self):
		self.tasks = []
		self.vservices = {}
		self.name = None
		import networkx as nx 
		self.graph = nx.DiGraph()

	def add(self, task):
		self.graph.add_node(task)
		self.tasks.append(task)

	def remove(self, task):
		self.tasks.remove(task)
		self.graph.remove_node(task)

	def add_edge(self, op1, op2):
		assert op1 in self.graph.nodes
		assert op2 in self.graph.nodes
		self.graph.add_edge(op1, op2)

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
						
	def set_costs(costs):
		for p, c in costs:
			self.cost[p] = c
					
	def get_costs(self):
		return self.costs

class Provider:
	def __init__(self, name):
		self.name = name
		self.region = None
		self.zone = None

	def set_region(self, region):
		self.region = region
	
	def set_zone(self, zone):
		self.zone = zone
		
providers = {}
filename = '/providers/__FILE_NAME_PROVIDERS__'
with open(filename, 'r') as file:
	for line in file:
		provider_name = line.strip().split(",")[0]
		pp = Provider(provider_name)
		pp.set_region(line.strip().split(",")[1])
		pp.set_zone(line.strip().split(",")[2])
		providers[provider_name] = pp
print(providers, end='')
print('\n'+'-'*5)

vservices = {}
vs_costs = {}
filename = '/vservices/__FILE_NAME_VSERVICES__'
with open(filename, 'r') as file:
	for line in file:
		vservice_name = line.strip().split(":")[0]
		vservice_provider = line.strip().split(":")[1].split(",")[0]
		vservice_cost = line.strip().split(":")[1].split(",")[1]
		if vservice_name not in vservices:
			vservices[vservice_name] = VService(vservice_name)
		vs_costs[vservice_provider] = vservice_cost
		vservices[vservice_name].set_costs(vs_costs)
print(vservice, end='')
print('\n'+'-'*5)

filename = '/tasks/__FILE_NAME_TASKS__'
tasks = {}
with open(filename, 'r') as file:
	for line in file:
		tt = line.split(":")[0]
		vservices = line.split(":")[1].split(",")
		tt = line.strip()
		task = Task(tt)
		tasks[tt] = task
		for vs in vservices:
			task.add_vservice(vservices[vs])
		dag.add(task)
		print(line, end='')
print('\n'+'-'*5)

filename = '/edges/__FILE_NAME_EDGES__'
with open(filename, 'r') as file:
	for line in file:
		print(line, end='')
print('\n'+'-'*5)

filename = '/providerattr/__FILE_NAME_PROVIDERATTR__'
with open(filename, 'r') as file:
	for line in file:
		print(line, end='')
print('\n'+'-'*5)



for pp in ['aws_east', 'gcp']:
	providers[pp] = Provider(pp)