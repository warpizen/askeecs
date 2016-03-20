var askeecsApp = angular.module('askeecs', ['angularMoment', 'ngRoute', 'askeecsControllers', 'ngCookies'])

askeecsApp.config(['$routeProvider',
  function($routeProvider) {
    $routeProvider.
      when('/page/:pageIdx', {
        templateUrl: 'partials/question-list.html',
        controller: 'QuestionListCtrl'
      }).
      when('/questions/:questionId', {
        templateUrl: 'partials/question-detail.html',
        controller: 'QuestionDetailCtrl'
      }).
      when('/ask', {
        templateUrl: 'partials/question-post.html',
        controller: 'QuestionAskCtrl'
      }).
      when('/update/:questionId', {
        templateUrl: 'partials/question-update.html',
        controller: 'QuestionUpdateCtrl'
      }).
      when('/register', {
        templateUrl: 'partials/register.html',
        controller: 'RegisterCtrl'
      }).
      when('/login', {
        templateUrl: 'partials/login.html',
        controller: 'LoginCtrl'
      }).
      otherwise({
        redirectTo: '/page/1'
      });
  }
]);

askeecsApp.run(function($rootScope, $location, AuthService, FlashService, 
  SessionService) {
  // how to work ? 
  var routesThatRequireAuth = ['/ask'];

  $rootScope.authenticated = SessionService.get('authenticated');
  $rootScope.user = SessionService.get('user');

  $rootScope.$on('$routeChangeStart', function (event, next, current) {
    FlashService.clear()
    if(_(routesThatRequireAuth).contains($location.path()) && 
      !AuthService.isLoggedIn()) {
      FlashService.show("Please login to continue");
      $location.path('/login');
    }
  });
});

askeecsApp.factory('logsOutUserOn401', ['$q', '$injector', 
  function($q, $injector) {
  var logsOutUserOn401 = {
    responseError: function(response) {
      if(response.status === 401) { // HTTP NotAuthorized
        var SessionService = $injector.get('SessionService');
        var $location = $injector.get('$location');
        var FlashService = $injector.get('FlashService');
        var $timeout = $injector.get('$timeout');
        SessionService.unset('authenticated')
        FlashService.show(response.data.Message);
        $timeout(function(){$location.path("/login");}, 1000);
        return $q.reject(response)
      } else {
        return $q.reject(response)
      }
    }
  };
  return logsOutUserOn401;
}]);

askeecsApp.config(['$httpProvider', function($httpProvider) {
  /*
  var logsOutUserOn401 = function ($location, $q, SessionService, FlashService) {
    var success = function (res) {
      return res;
    }
    var error   = function (res) {
      if(res.status === 401) { // HTTP NotAuthorized
        SessionService.unset('authenticated')
        FlashService.show(res.data.Message);
        $location.path("/login");
        return $q.reject(res)
      } else {
        return $q.reject(res)
      }
    }
    return function(promise) {
      return promise.then(success, error)
    }
  }
  */
  //$httpProvider.responseInterceptors.push(logsOutUserOn401);
  $httpProvider.interceptors.push('logsOutUserOn401');
}]);

// copyed from 
// http://stackoverflow.com/questions/3964710/replacing-selected-text-in-the-textarea
askeecsApp.factory("ToolbarService", function () {
  var getInputSelection = function(el) {
    var start = 0, end = 0, normalizedValue, range, 
      textInputRange, len, endRange;
    if (typeof el.selectionStart == "number" && 
      typeof el.selectionEnd == "number") {
      start = el.selectionStart;
      end = el.selectionEnd;
    } else {
      range = document.selection.createRange();

      if (range && range.parentElement() == el) {
        len = el.value.length;
        normalizedValue = el.value.replace(/\r\n/g, "\n");

        // Create a working TextRange that lives only in the input
        textInputRange = el.createTextRange();
        textInputRange.moveToBookmark(range.getBookmark());

        // Check if the start and end of the selection are at the very end
        // of the input, since moveStart/moveEnd doesn't return what we want
        // in those cases
        endRange = el.createTextRange();
        endRange.collapse(false);

        if (textInputRange.compareEndPoints("StartToEnd", endRange) > -1) {
          start = end = len;
        } else {
          start = -textInputRange.moveStart("character", -len);
          start += normalizedValue.slice(0, start).split("\n").length - 1;

          if (textInputRange.compareEndPoints("EndToEnd", endRange) > -1) {
            end = len;
          } else {
            end = -textInputRange.moveEnd("character", -len);
            end += normalizedValue.slice(0, end).split("\n").length - 1;
          }
        }
      }
    }

    return {
        start: start,
        end: end
    };
  };
  return {
    replaceSelectedText: function(el, text) {
      if (el == null) return;
      var sel = getInputSelection(el), val = el.value;
      return val.slice(0, sel.start) + text + val.slice(sel.end);
    },
    append: function(el, textfront, textback) {
      if (el == null) return;
      var sel = getInputSelection(el), val = el.value;
      var org = val.slice(sel.start, sel.end);
      return val.slice(0, sel.start) + textfront + org + textback + 
        val.slice(sel.end);
    },
    appendFront: function(el, text) {
      if (el == null) return;
      var sel = getInputSelection(el), val = el.value;
      var org = val.slice(sel.start, sel.end);
      var splited = org.split('\n');
      org = "";
      for (var i in splited) {
        if (i == splited.length) {
          org += text + splited[i];
        } else {
          org += text + splited[i] + "\n";
        }
      }
      return val.slice(0, sel.start) + org + val.slice(sel.end); 
    },
    appendBoth: function(el, text) {
      if (el == null) return;
      var sel = getInputSelection(el), val = el.value;
      var org = val.slice(sel.start, sel.end);
      var splited = org.split('\n');
      org = "";
      for (var i in splited) {
        if (i == splited.length) {
          org += text + splited[i] + text;
        } else {
          org += text + splited[i] + text + "\n";
        }
      }
      return val.slice(0, sel.start) + org + val.slice(sel.end); 
    },
    appendBack: function(el, text) {
      if (el == null) return;
      var sel = getInputSelection(el), val = el.value;
      var org = val.slice(sel.start, sel.end);
      var splited = org.split('\n');
      org = "";
      for (var i in splited) {
        if (i == splited.length) {
          org += splited[i] + text;
        } else {
          org += splited[i] + text + "\n";
        }
      }
      return val.slice(0, sel.start) + org + val.slice(sel.end); 
    },
    md2Html: function() {
      var src = $scope.markdown || ""
      $scope.html = $window.marked(src);
      $scope.htmlSafe = $sce.trustAsHtml($scope.html);
    }

  }
});

/*
askeecsApp.factory("SessionService", function () {
  return {
    get: function (key) {
      return sessionStorage.getItem(key);
    },
    set: function (key, val) {
      return sessionStorage.setItem(key, val);
    },
    unset: function (key) {
      return sessionStorage.removeItem(key);
    }
  }
});
*/

askeecsApp.factory("SessionService", function () {
  return {
    get: function (key) {
      return JSON.parse(localStorage.getItem(key));
    },
    set: function (key, val) {
      return localStorage.setItem(key, JSON.stringify(val));
    },
    unset: function (key) {
      return localStorage.removeItem(key);
    }
  }
});

askeecsApp.factory("AuthService", ['$rootScope', '$http', '$location', 'SessionService', 'FlashService',
  function($rootScope, $http, $location, SessionService, FlashService) {
    var cacheSession = function (user) {
      SessionService.set('authenticated', true);
      SessionService.set('user', user);
      $rootScope.authenticated = true;
      $rootScope.user = user;
    }
    var uncacheSession = function () {
      SessionService.unset('authenticated');
      SessionService.unset('user');
      $rootScope.authenticated = false;
      $rootScope.user = {};
    }
    var loginError = function (res) {
      FlashService.show(res.Message);
    }
    var protect = function (secret, salt) {
      var SHA256 = new Hashes.SHA256;
      return SHA256.hex(salt + SHA256.hex(secret));
    }
    var hash = function () {
      var s = ""
      var SHA256 = new Hashes.SHA256;
      for ( var i = 0; i < arguments.length; i++) {
        s += arguments[i];
      }
      return SHA256.hex(s);
    }
    return {
      login: function (credentials, fn) {
        // Friendly vars
        var u = credentials.Username;
        var p = credentials.Password;

        credentials.Username = "";
        credentials.Password = "";

        // Get a salt for this session
        $http.post("/register/salt", {"Username" : u})
          .success(function(user_salt) {
            $http.post("/salt", {"Username" : u})
              .success(function(session_salt) {

                // Produce the "Password" to send
                p = protect (u + p, user_salt.Salt);
                p = hash( p , session_salt)

                // Try to login
                var login = $http.post("/login", 
                  {"Username": u, "Password": p, "Salt": session_salt});

                login.success(cacheSession);
                login.success(FlashService.clear);
                login.error(loginError);

                if ( typeof fn === "function" )
                  login.success(fn);
              }
            )
          }
        )
      },
      logout: function (fn) {
        var logout =  $http.post("/logout");
        logout.success(uncacheSession);
        if ( typeof fn === "function" )
          logout.success(fn);
      },
      register: function (credentials, fn) {
        // Friendly vars
        var u = credentials.Username;
        var p = credentials.Password;
        credentials.Username = "";
        credentials.Password = "";
        var s = protect(Date.now(), Math.random());
        // Produce the "Password" to send
        p = protect(u + p, s);
        var register = $http.post("/register", 
            {"Username" : u, "Password" : p, "Salt" : s });
        if ( typeof fn === "function")
          register.success(fn);
      },
      isLoggedIn: function () {
        return SessionService.get('authenticated');
      },
      currentUser: function () {
        if ( this.isLoggedIn() ) {
          return SessionService.get('user');
        }
        return {};
      }
    }
  }
]);

askeecsApp.factory("FlashService", function ($rootScope) {
  return {
    show: function (msg) {
      $rootScope.flashn = 1;
      $rootScope.flash = msg
    },
    clear: function () {
      if ( $rootScope.flashn-- == 0 )
        $rootScope.flash = ""
    }
  }
});

askeecsApp.factory('Questions', ['$http',
  function ($http) {
    var urlBase = '/q'
    var store = []
    var f   = {};
    var p = function (data) {
      this.success = function (fn) {
        fn(data)
      }
    }
    f.List = function(page) {
      return $http.get(urlBase+'?page='+page)
        .success(function (data) {
          store = data;
        });
    }
    f.GetMaxPage = function() {
      return $http.get("/maxpage")
        .success(function (data) {
          // do nothing
        });
    }
    f.TrendingTags = function() {
      return $http.get("/trendingtags")
        .success(function(data) {
          //console.log(data);
        });
    }
    f.Get = function (id, force) {
      if ( !force ) {
        for ( var i = 0; i < store.length; i++ ) {
          if ( store[i].ID == id )
            return new p(store[i]); 
        }
      }
      return $http.get(urlBase + '/' + id);
    }
    f.Insert = function (item) {
      return $http.post(urlBase, item)
        .success(function(data) {
          store.push(data);
        });
    }
    f.Update = function (item) {
      return $http.put(urlBase + '/' + item.ID, item)
        .success(function (data) {
          for ( var i = 0; i < store.length; i++ ) {
            if ( store[i].ID == id )
              return store[i] = data;
          }
        });
    }
    f.Delete = function (id) {
      return $http.delete(urlBase + '/' + id)
        .success(function (data) {
          for ( var i = 0; i < store.length; i++ ) {
            if ( store[i].ID == id )
              return store.splice(i, 1);
          }
        })
    }
    return f;
  }
]);

askeecsApp.directive('askeecsLogout', function (AuthService) {
  return {
    restrict: 'A',
    link: function(scope, element, attrs) {
    var evHandler = function(e) {
      e.preventDefault;
      AuthService.logout();
      return false;
    }
    element.on ? element.on('click', evHandler) : 
      element.bind('click', evHandler);
    }
  }
});

askeecsApp.directive('question', ['Questions',
  function (Questions) {
    function link ( scope, element, attributes ) {
      console.log("Generating question...", attributes.question);
      Questions.Get(attributes.question).success(function(data) {
        console.log(data)
      })
    }
    return {
      restruct: 'A',
      link: link
    }
  }
]);

askeecsApp.directive('comment', ['Questions',
  function (Questions) {
    function link ( scope, element, attributes ) {
      console.log("Generating comment...", attributes.question);
      Questions.Get(attributes.question).success(function(data) {
        console.log(data)
      })
    }
    return {
      restruct: 'A',
      link: link
    }
  }
]);

askeecsApp.directive('markedbox', function() {
  return {
    restrict: 'E',
    templateUrl: 'partials/markedbox.html',
    controller: 'MarkedBoxCtrl'
  }
});

askeecsApp.directive('modaldel', function () {
  return {
    template: 
      '<div class="modal fade">' + 
        '<div class="modal-dialog">' + 
          '<div class="modal-content">' + 
            '<div class="modal-header">' + 
              '<button type="button" class="close"' + 
              '   data-dismiss="modal" aria-hidden="true">&times;</button>' + 
              '<h4 class="modal-title">{{ title }}</h4>' + 
            '</div>' + 
            '<div class="modal-body" ng-transclude></div>' + 
          '</div>' + 
        '</div>' + 
      '</div>',
    restrict: 'E',
    transclude: true,
    replace:true,
    scope:true,
    link: function postLink(scope, element, attrs) {
      scope.title = attrs.title;

      scope.$watch(attrs.visible, function(value){
        if(value == true)
          $(element).modal('show');
        else
          $(element).modal('hide');
      });

      $(element).on('shown.bs.modal', function(){
        scope.$apply(function(){
          scope.$parent[attrs.visible] = true;
        });
      });

      $(element).on('hidden.bs.modal', function(){
        scope.$apply(function(){
          scope.$parent[attrs.visible] = false;
        });
      });
    }
  };
});

askeecsApp.filter('commentremark', function () {
  return function(input) {
    if(input === 0)
      return "at least enter 15 characters";
    else if(input < 15)
      return "" + 15 - input + " more to go..."
    else
      return 600 - input + " characters left"
  }
});
