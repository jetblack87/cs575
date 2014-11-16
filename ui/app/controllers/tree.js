(function() {
	var app, deps;

	deps = ['angularBootstrapNavTree'];

	if (angular.version.full.indexOf('1.2') >= 0) {
		deps.push('ngAnimate');
	}

	app = angular.module('treeApp', deps);

	app.controller('treeController', function($scope, $timeout) {
		var apple_selected, tree, treedata_avm;
		$scope.my_tree_handler = function(branch) {
			var _ref;
			$scope.output = 'You selected: ' + branch.label;
			if ((_ref = branch.data) != null ? _ref.description : void 0) {
				return $scope.output += '(' + branch.data.description + ')';
			}
		};
		apple_selected = function(branch) {
			return $scope.output = 'APPLE! : ' + branch.label;
		};

		treedata_avm = [
		                {
		                	label: 'Animal',
		                	children: [
		                	           {
		                	        	   label: 'Dog',
		                	        	   data: {
		                	        		   description: 'A man\'s best friend'
		                	        	   }
		                	           },
		                	           {
		                	        	   label: 'Cat',
		                	        	   data: {
		                	        		   description: 'Felis catus'
		                	        	   }
		                	           },
		                	           {
		                	        	   label: 'Fish',
		                	        	   data: {
		                	        		   descrption: 'Glub glub!'
		                	        	   }
		                	           }
		                	           ]
		                },
		                {
		                	label: 'Vegetable',
		                	data: {
		                		definition: 'A plan or part of a plant...',
		                		data_can_contain_anything: true
		                	},
		                	onSelect: function(branch) {
		                		return $scope.output = 'Vegetable: ' + branch.data.definition;
		                	},
		                	children: [
		                	           {
		                	        	   label: 'Carrot',
		                	        	   data: {
		                	        		   description: 'Good for your eyes!'
		                	        	   }
		                	           },
		                	           {
		                	        	   label: 'Broccoli',
		                	        	   data : {
		                	        		   description: 'Good for your children'
		                	        	   }
		                	           }
		                	           ]
		                }
		                ];
		$scope.my_data = treedata_avm;
		$scope.my_tree = tree = {};

	});


}).call(this);