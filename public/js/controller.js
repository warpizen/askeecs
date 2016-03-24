var askeecsControllers = angular.module('askeecsControllers', ['ngCookies']);

askeecsControllers.controller('QuestionListCtrl', 
  ['$scope', 'Questions', '$routeParams', '$rootScope',
  function ($scope, Questions, $routeParams, $rootScope) {
    //
    $rootScope.trending = {};
    Questions.GetMaxPage().success(function(count) {
      $scope.maxPage = count;
      var requestPage = $scope.page = parseInt($routeParams.pageIdx);
      $scope.prevPage = requestPage - 1;
      $scope.nextPage = requestPage + 1;

      if (requestPage < 1) {
        $scope.page = 1;
      } else if (requestPage > $scope.maxPage) {
        $scope.page = maxPage;
      }
      // always asynchromous $http?
      Questions.List($scope.page)
        .success(function (questions) {
          $scope.questions = questions;
        })
        .error(function (error) {
          $scope.error = 'Unable to load questions';
        });
      // Get Trending Tags
      Questions.TrendingTags()
        .success(function(data) {
          //console.log(data);
          $rootScope.trending.tags = data;
        });
    });
  }
]);

askeecsControllers.controller('RegisterCtrl', 
  ['$scope', '$http', '$location', 'AuthService',
  function ($scope, $http, $location, AuthService) {
    var credentials = { "Username": "", "Password": "", "CPassword": "" }
    $scope.credentials = credentials; 
    $scope.processForm = function () {
      // Make sure they have entered a password that matches
      if($scope.credentials.Password != $scope.credentials.CPassword) {
        console.log("Missed matched password");
        return;
      }
      // We don't need this to be passed along
      delete $scope.credentials.CPassword;
      // Register the user and redirect them to the login page
      AuthService.register($scope.credentials, function () {
        $location.path("/login");
      });
      // Make sure we wipe out the credentials
      $scope.credentials = credentials; 
    }
  }
]);

askeecsControllers.controller('LoginCtrl', 
  ['$scope', '$http', '$location', 'AuthService',
  function ($scope, $http, $location, AuthService) {
    var credentials = { "Username": "", "Password": "", "Salt": "" }
    $scope.credentials = credentials
    $scope.processForm = function () {
      // Log the user in and direct them tot he home page
      AuthService.login($scope.credentials, function () {
        $location.path("/");
      });
      // Make sure we wipe out the credentials
      $scope.credentials = credentials
    }
  }
]);

askeecsControllers.controller('QuestionAskCtrl', 
  ['$scope', '$http', '$window', '$sce', '$location', 'ToolbarService',
  function ($scope, $http, $window, $sce, $location, ToolbarService) {
    var question = {"title" : "", "tags" : ""}
    $scope.question = question;

    $scope.processForm = function () {
      // Remove any previous error statements
      $scope.error = {}
      // Default to a non error state
      var err = false;
      if ($scope.markdown.length < 50) {
        $scope.error.markdown = "Your question must be 50 characters or more."
        err = true;
      }
      if ($scope.question.title.length == 0) {
        $scope.error.title = "You must enter a title."
        err = true;
      }
      if ($scope.question.tags.length == 0) {
        $scope.error.tags = "You must have at least one tag."
        err = true;
      }
      if (err) {
        return;
      }
      $http({
        method: 'POST',
        url: '/q',
        data: { 
          Title:$scope.question.title, 
          Body: $scope.markdown, 
          Tags: $scope.question.tags.split(' ')
        }
      }).success(function(data) {
        // TODO: this should be a JSON response
        $location.path("/questions/"+data); 
      });
      // TODO: Failure
    }
  }
]);

askeecsControllers.controller('QuestionUpdateCtrl', 
  ['$scope', '$http', '$window', '$sce', '$location', 'ToolbarService',
  '$routeParams', 
  function ($scope, $http, $window, $sce, $location, ToolbarService, 
    $routeParams) {

    var question = {"title" : "", "tags" : ""}
    $scope.question = question;

    $http.get('/q/' + $routeParams.questionId).success(function(data) {
      $scope.markdown = data.Body;
      $scope.question.title = data.Title;
      $scope.question.tags = String(data.Tags).replace(",", " ");
      $scope.md2Html();
    });


    $scope.processForm = function () {
      // Remove any previous error statements
      $scope.error = {}
      // Default to a non error state
      var err = false;
      if ($scope.markdown.length < 50) {
        $scope.error.markdown = "Your question must be 50 characters or more."
        err = true;
      }
      if ($scope.question.title.length == 0) {
        $scope.error.title = "You must enter a title."
        err = true;
      }
      if ($scope.question.tags.length == 0) {
        $scope.error.tags = "You must have at least one tag."
        err = true;
      }
      if (err) {
        return;
      }
      $http({
        method: 'POST',
        url: '/q/' + $routeParams.questionId,
        data: { 
          Title:$scope.question.title, 
          Body: $scope.markdown, 
          Tags: $scope.question.tags.split(' ')
        }
      }).success(function(data) {
        // TODO: this should be a JSON response
        $location.path("/questions/"+data.ID); 
      });
      // TODO: Failure
    }
  }
]);

askeecsControllers.controller('QuestionDetailCtrl', 
  ['$scope', '$routeParams', '$http', '$window', '$sce', 'ToolbarService', 
  '$document', 'FlashService', '$location', 'Questions', '$rootScope',
  function ($scope, $routeParams, $http, $window, $sce, ToolbarService, 
    $document, FlashService, $location, Questions, $rootScope) {
    $rootScope.trending = {}
    $scope.comment = { "Body" : "" };
    $scope.response = { "Body" : "" };
    $http.get('/q/' + $routeParams.questionId).success(function(data) {
      $scope.question = data;
      $scope.question.HTML = $sce.trustAsHtml($window.marked(data.Body));
      // for ??
      for (var i in data.Responses) {
        $scope.question.Responses[i].HTML = 
          $sce.trustAsHtml($window.marked(data.Responses[i].Body));
      }
      console.log(data)
    });
    // Get Trending Tags
    Questions.TrendingTags()
      .success(function(data) {
        //console.log(data);
        $rootScope.trending.tags = data;
      });
    $scope.voteUp = function () {
      $http({
        method: 'GET',
        url: '/q/' + $scope.question.ID + '/vote/up',
        data: {}
      }).success(function(data) {
        $scope.question.Upvotes = data.Upvotes
      });
    }
    $scope.voteDown = function () {
      $http({
        method: 'GET',
        url: '/q/' + $scope.question.ID + '/vote/down',
        data: {}
      }).success(function(data) {
        $scope.question.Downvotes = data.Downvotes
      });
    }

    //$scope.markdown="";

    // Can a comment have this own controller and it's own scope?
    $scope.processComment = function () {
      delete $scope.errorComment;
      var err = false;
      if ( $scope.comment.Body.length < 15 ) {
        $scope.errorComment = "Your comment must be at least 15 characters"
        err = true;
      }
      if (err) return;
      $http({
        method: 'POST',
        url: '/q/' + $scope.question.ID + '/comment',
        data: $scope.comment
      }).success(function(data) {
        delete $scope.comment_add;
        delete $scope.comment.Body;
        $scope.question.Comments.push(data);
      });
    }
    // 
    $scope.updateComment = function (body, cid) {
      delete $scope.errorComment;
      if (body.length != 0) {
        $scope.comment.Body = body;
      }
      var err = false;
      if ( $scope.comment.Body.length < 15 ) {
        $scope.errorComment = "Your comment must be at least 15 characters"
        err = true;
      }
      if (err) return;
      $http({
        method: 'POST',
        url: '/q/' + $scope.question.ID + '/comment/' + cid,
        data: $scope.comment
      }).success(function(data) {
        delete $scope.comment_update;
        delete $scope.comment.Body;
        for (var i in $scope.question.Comments) {
          if ($scope.question.Comments[i].ID == data.ID) {
            $scope.question.Comments[i] = data; 
            break;
          }
        }
      });
    }
    $scope.deleteQuestionComment = function(cid) {
      delete $scope.errorComment;
      $http({
        method: 'PUT',
        url: '/q/' + $scope.question.ID + '/comment/' + cid,
        data: {}
      }).then(function success(r) {
        delete $scope.comment_delete;
        for (var i in $scope.question.Comments) {
          if ($scope.question.Comments[i].ID == r.data.ID) {
            $scope.question.Comments.splice(i, 1);
            break;
          }
        }
      }, function failure(r) {
        FlashService.show(r.data);
      });
    }
    $scope.deleteAnswer = function(rid) {
      delete $scope.errorComment;
      $http({
        method: 'PUT',
        url: '/q/' + $scope.question.ID + '/response/' + rid,
        data: {}
      }).then(function success(r) {
        delete $scope.answer_delete;
        for (var i in $scope.question.Responses) {
          if ($scope.question.Responses[i].ID == r.data.ID) {
            $scope.question.Responses.splice(i, 1);
            break;
          }
        }
      }, function failure(r) {
        FlashService.show(r.data);
      });
    }
    $scope.deleteQuestion = function() {
      delete $scope.errorComment;
      $http({
        method: 'PUT',
        url: '/q/' + $scope.question.ID,
        data: {}
      }).then(function success(r) {
        delete $scope.question_delete;
        $location.path("/page/1"); 
      }, function failure(r) {
        FlashService.show(r.data);
      });
    }
    $scope.processForm = function (externId) {
      console.log($scope.response.Body);
      delete $scope.error.markdown;
      var err = false;
      if ($scope.markdown.length < 50) {
        $scope.error.markdown = "Your response must be 50 characters or more."
        err = true;
      }
      if (err) {
        return;
      }
      $scope.response.Body = $scope.markdown;
      if (externId == "" || typeof externId == "undefined") {
        $http({
          method: 'POST',
          url: '/q/' + $scope.question.ID + '/response',
          data: $scope.response
        }).success(function(data) {
          $scope.extern = false;
          $scope.markdown = "";
          $scope.md2Html();
          $scope.question.Responses.push(data);
          var i = $scope.question.Responses.length;
          $scope.question.Responses[i-1].HTML = 
            $sce.trustAsHtml($window.marked(data.Body));
        });
      } else {
        $http({
          method: 'POST',
          url: '/q/' + $scope.question.ID + '/response/' + externId,
          data: $scope.response
        }).success(function(data) {
          $scope.extern = false;
          $scope.markdown = "";
          $scope.md2Html();
          for (var i in $scope.question.Responses) {
            if ($scope.question.Responses[i].ID == data.ID) {
              $scope.question.Responses[i] = data
              $scope.question.Responses[i].HTML = 
                $sce.trustAsHtml($window.marked(data.Body));
              break;
            }
          }
        });
      }
    }
  }
]);

askeecsControllers.controller('MarkedBoxCtrl', 
  ['$scope', '$window', '$sce', 'ToolbarService', '$document',
  function ($scope, $window, $sce, ToolbarService, $document) {
    $scope.markdown = "";
    $scope.extern = false;
    $scope.externId = "";
    $scope.putData = function(data, id) {
      if (data.length != 0) {
        $scope.markdown = data;
        $scope.externId = id;
        $scope.extern = true;
      }
      $scope.md2Html();
    }
    $scope.md2Html = function() {
      var src = $scope.markdown || ""
      $scope.html = $window.marked(src);
      $scope.htmlSafe = $sce.trustAsHtml($scope.html);
    }
    $scope.toolbarBold = function() {
      console.log("toolbar bold called");
      var el = document.querySelector('#mk-input');
      $scope.markdown = ToolbarService.appendBoth(el, "__");
      $scope.md2Html();
    }
    $scope.toolbarItalic = function() {
      console.log("toolbar Italic called");
      var el = document.querySelector('#mk-input');
      $scope.markdown = ToolbarService.appendBoth(el, "_");
      $scope.md2Html();
    }
    $scope.toolbarListul = function() {
      console.log("toolbar list-ul called");
      var el = document.querySelector('#mk-input');
      $scope.markdown = ToolbarService.appendFront(el, "* ");
      $scope.md2Html();
    }
    $scope.toolbarListol = function() {
      console.log("toolbar list-ol called");
      var el = document.querySelector('#mk-input');
      $scope.markdown = ToolbarService.appendFront(el, "1. ");
      $scope.md2Html();
    }
    $scope.toolbarQuote = function() {
      var el = document.querySelector('#mk-input');
      $scope.markdown = ToolbarService.appendFront(el, "> ");
      $scope.md2Html();
    }
    $scope.toolbarCode = function() {
      var el = document.querySelector('#mk-input');
      $scope.markdown = ToolbarService.append(el, "```code\n", "\n```");
      $scope.md2Html();
    }


  }
]);
