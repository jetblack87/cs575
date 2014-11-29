(function() {
	var app, deps;

	deps = [ 'ngRoute', 'angularBootstrapNavTree' ];

	if (angular.version.full.indexOf('1.2') >= 0) {
		deps.push('ngAnimate');
	}

	app = angular.module('treeApp', deps);

	app.config(function($routeProvider) {
		$routeProvider.when('/', {
			controller : 'TreeController',
			templateUrl : '../views/partials/splash.html'
		}).when('/domains/:maestroNodeKey', {
			controller : 'DomainsController',
			templateUrl : '../views/partials/domains.html'
		}).when('/agents/:maestroNodeKey', {
			controller : 'DummyDetailsController',
			templateUrl : '../views/partials/details.html'
		}).when('/processes/:maestroNodeKey', {
			controller : 'ProcessesController',
			templateUrl : '../views/partials/processes.html'
		}).otherwise({
			redirectTo : '/'
		});
	});

	app.controller('TreeController', function($scope) {
		var tree, treedata_avm, mock_maestro_data;
		$scope.my_tree_handler = function(branch) {
			var _ref;
			$scope.output = 'You selected: ' + branch.label;
			if ((_ref = branch.data) != null ? _ref.description : void 0) {
				return $scope.output += ' (' + branch.data.description + ')';
			}
		};

		treedata_avm = [ {
			label : 'Animal',
			children : [ {
				label : 'Dog',
				data : {
					description : 'A man\'s best friend'
				}
			}, {
				label : 'Cat',
				data : {
					description : 'Felis catus'
				}
			}, {
				label : 'Fish',
				data : {
					descrption : 'Glub glub!'
				}
			} ]
		}, {
			label : 'Vegetable',
			data : {
				definition : 'A plan or part of a plant...',
				data_can_contain_anything : true
			},
			onSelect : function(branch) {
				return $scope.output = 'Vegetable: ' + branch.data.definition;
			},
			children : [ {
				label : 'Carrot',
				data : {
					description : 'Good for your eyes!'
				}
			}, {
				label : 'Broccoli',
				data : {
					description : 'Good for your children'
				}
			} ]
								} ];

						mock_maestro_data = [ {
							"Name" : "d01",
							"Key" : "L21hZXN0cm8vZDAx",
							"Runtime" : {
								"Agents" : [
										{
											"Name" : "a01",
											"Key" : "L21hZXN0cm8vZDAxL3J1bnRpbWUvYWdlbnRzL2EwMQ==",
											"AgentClass" : "",
											"OS" : "",
											"Processes" : [ {
												"Name" : "p01",
												"Key" : "L21hZXN0cm8vZDAxL3J1bnRpbWUvYWdlbnRzL2EwMS9wcm9jZXNzZXMvcDAx",
												"Command" : "C:/Windows/notepad.exe",
												"Arguments" : "",
												"ProcessClass" : "",
												"AdminState" : "on",
												"OperState" : "on",
												"Pid" : 6116
											} ]
										},
										{
											"Name" : "a02_linux",
											"Key" : "L21hZXN0cm8vZDAxL3J1bnRpbWUvYWdlbnRzL2EwMl9saW51eA==",
											"AgentClass" : "",
											"OS" : "",
											"Processes" : [ {
												"Name" : "p02_linux",
												"Key" : "L21hZXN0cm8vZDAxL3J1bnRpbWUvYWdlbnRzL2EwMl9saW51eC9wcm9jZXNzZXMvcDAyX2xpbnV4",
												"Command" : "/bin/sleep",
												"Arguments" : "1000",
												"ProcessClass" : "",
												"AdminState" : "on",
												"OperState" : "on",
												"Pid" : 38274
											} ]
										} ]
							},
							"Config" : {
								"Agents" : [
										{
											"Name" : "a01",
											"Key" : "L21hZXN0cm8vZDAxL2NvbmZpZy9hZ2VudHMvYTAx",
											"AgentClass" : "",
											"OS" : "",
											"Processes" : [ {
												"Name" : "p01",
												"Key" : "L21hZXN0cm8vZDAxL2NvbmZpZy9hZ2VudHMvYTAxL3Byb2Nlc3Nlcy9wMDE=",
												"Command" : "",
												"Arguments" : "",
												"ProcessClass" : "/maestro/d01/config/processes/p01",
												"AdminState" : "",
												"OperState" : "",
												"Pid" : 0
											} ]
										},
										{
											"Name" : "a02_linux",
											"Key" : "L21hZXN0cm8vZDAxL2NvbmZpZy9hZ2VudHMvYTAyX2xpbnV4",
											"AgentClass" : "",
											"OS" : "",
											"Processes" : [ {
												"Name" : "p02_linux",
												"Key" : "L21hZXN0cm8vZDAxL2NvbmZpZy9hZ2VudHMvYTAyX2xpbnV4L3Byb2Nlc3Nlcy9wMDJfbGludXg=",
												"Command" : "",
												"Arguments" : "",
												"ProcessClass" : "/maestro/d01/config/processes/p02_linux",
												"AdminState" : "",
												"OperState" : "",
												"Pid" : 0
											} ]
										} ],
								"Processes" : [
										{
											"Name" : "p02_linux",
											"Key" : "L21hZXN0cm8vZDAxL2NvbmZpZy9wcm9jZXNzZXMvcDAyX2xpbnV4",
											"Command" : "/bin/sleep",
											"Arguments" : "1000",
											"ProcessClass" : "",
											"AdminState" : "",
											"OperState" : "",
											"Pid" : 0
										},
										{
											"Name" : "p01",
											"Key" : "L21hZXN0cm8vZDAxL2NvbmZpZy9wcm9jZXNzZXMvcDAx",
											"Command" : "C:/Windows/notepad.exe",
											"Arguments" : "",
											"ProcessClass" : "",
											"AdminState" : "",
											"OperState" : "",
											"Pid" : 0
										} ]
							}
						} ];

		$scope.domains = mock_maestro_data;
		$scope.my_tree = tree = {};
	});

	app.controller('DummyDetailsController', function($scope, $routeParams) {
		// Dummy data - details for p01
		$scope.details = {
			"Name" : "p01",
			"Key" : "L21hZXN0cm8vZDAxL2NvbmZpZy9wcm9jZXNzZXMvcDAx",
			"Command" : "C:/Windows/notepad.exe",
			"Arguments" : "",
			"ProcessClass" : "",
			"AdminState" : "on",
			"OperState" : "on",
			"Pid" : 0
		};

		$scope.maestroNode = $routeParams.maestroNodeKey;
	});
	
	app.controller('DomainsController', function($scope, $http, $routeParams) {
		var urlPrefix = "http://tux64-11.cs.drexel.edu:8080/domains/";
		var url = urlPrefix + $routeParams.maestroNodeKey;
		
		$http.get(url).success(function(data) {
			$scope.details = data;
		});
	});
	
	app.controller('ProcessesController', function($scope, $http, $routeParams) {
		var urlPrefix = "http://tux64-11.cs.drexel.edu:8080/processes/";
		var url = urlPrefix + $routeParams.maestroNodeKey;
		
		$scope.getProcessData = function() {
			$http.get(url).success(function(data) {
		
			$scope.details = data;
		})};
		
		$scope.startProcess = function() {
			var startData = '{"AdminState": "on"}';
			
			$http.patch(url, startData).success(function(data) {
				$scope.details = data;
			});
		};
		
		$scope.stopProcess = function() {
			var stopData = '{"AdminState": "off"}';
			
			$http.patch(url, stopData).success(function(data) {
				$scope.details = data;
			});
		};
		
		// Actually get the process data when the controller first runs
		$scope.getProcessData();
		
	});

}).call(this);