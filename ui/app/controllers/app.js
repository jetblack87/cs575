(function() {
	var app, deps;

	deps = [ 'ngRoute' ];

	app = angular.module('treeApp', deps);

	app.config(function($routeProvider) {
		$routeProvider.when('/', {
			controller : 'TreeController',
			templateUrl : '../views/partials/splash.html'
		}).when('/domains/:maestroNodeKey', {
			controller : 'DomainsController',
			templateUrl : '../views/partials/domains.html'
		}).when('/processes/:maestroNodeKey', {
			controller : 'ProcessesController',
			templateUrl : '../views/partials/processes.html'
		}).otherwise({
			redirectTo : '/'
		});
	});

	app.controller('TreeController', function($scope, $http) {
		var url = "http://tux64-11.cs.drexel.edu:8080/domains/";
		$http.get(url).success(function(data) {
			$scope.domains = data;
			$scope.my_tree = tree = {};
		});
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