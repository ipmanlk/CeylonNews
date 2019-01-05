var api = "https://pk.navinda.xyz/api/ceylon_news/v2.2/";
// store news list, posts temp
var newsList = {};
var newsPosts = {};
var newsSources = {};
var selectedSource = "null";

// current location in app
var currentPage = "news-list";
// current reading post
var currentPostId;

ons.ready(function() {
  // check disclamer notice
  if (!localStorage.getItem('showNotice')) {
    showNotice();
  }
  // disable built in back button handler of onsen
  ons.disableDeviceBackButtonHandler();
  // get news list
  showToast("Loading posts...");
  getNewsList("null", "null", "normal");
  // check for new articles
  setInterval(checkNewPosts, 60000);
});

// api requests
function getPostOnline(postId) {
  showToast("Loading post...");
  $.post(api, {
      action: "news_post",
      post_id: postId
    }, null, 'json')
    .done(function(data) {
      newsPosts[data.id] = data;
      showPost(postId, data);
      hideToast();
    })
    .fail(function() {
      hideToast();
      ons.notification.alert("Unable to load post!");
    });
}

function getSources() {
  $.post(api, {
      action: "sources_list"
    }, null, 'json')
    .done(function(data) {
      for (var item in data) {
        $('#menu-sources').append('<ons-list-item tappable onclick="loadSource(\'' + data[item].id + '\');">' + data[item].source + '</ons-list-item>');
        // add to object for later use
        newsSources[data[item].id] = data[item].source;
      }
    })
    .fail(function() {
      ons.notification.alert("Unable to read sources!");
    });
}

function getNewsList(postId, sourceId, mode) {
  $.post(api, {
      action: "news_list",
      post_id: postId,
      source_id: sourceId,
      mode: mode
    }, null, 'json')
    .done(function(data) {
      if (!jQuery.isEmptyObject(data)) {
        for (var item in data) {
          newsList[(data[item].id)] = data[item];
          if (mode == "normal") {
            $('#news-list-content').append(getNewListItem(data[item]));
          } else {
            $('#news-list-content').prepend(getNewListItem(data[item]));
          }
        }

        if (mode == "check") {
          showToast("New articles are available!");
          setTimeout(hideToast, 4000);
        } else {
          hideToast();
        }

        $('#load-more-btn').fadeIn();

      } else {
        if (mode == "normal") $('#load-more-btn').hide();
      }

      if (Object.keys(newsSources).length == 0) {
        getSources();
      }

    })
    .fail(function() {
      ons.notification.alert("Unable to read feeds!");
    });
}

// other functional tasks
function loadPost(postId) {
  if (postId in newsPosts) {
    getPostOffline(postId);
  } else {
    getPostOnline(postId);
  }
  currentPage = "post";
  currentPostId = postId;
}

function getPostOffline(postId) {
  showPost(postId, newsPosts[postId]);
}

function loadMoreNews() {
  // load more posts
  $('#load-more-btn').hide();
  showToast("Loading more posts...");
  var keys = Object.keys(newsList);
  var oldestId = keys[0];
  getNewsList(oldestId, selectedSource, "normal");
}

function refreshData() {
  if (currentPage == "news-list") {
    $('#load-more-btn').hide();
    $('#news-list-content').empty();
    showToast("Loading posts...");
    getNewsList("null", "null", "normal");
  } else if (currentPage == "post") {
    getPostOnline(currentPostId);
  }
}

function loadSource(sourceId) {
  var sourceName = newsSources[sourceId];
  selectedSource = sourceId;
  showToast("Loading posts...");
  $('#load-more-btn').hide();
  $('#news-list-content').empty();
  getNewsList("null", sourceId, "normal");
  $('#toolbar-title').text(sourceName);
  menu.close();
}

// element generators
function getNewListItem(post) {
  // generate li element for list
  var id, source, datetime, title, mainImg;
  id = post.id;
  source = post.source;
  datetime = post.datetime;
  title = fixEscapedHtml(post.title);
  mainImg = post.mainImg;
  var html = '<li id="' + id + '" class="list-item"><div class="list-item__left"><img class="list-item__thumbnail" src="' + mainImg + '" alt="mainImg"  onerror="fixBrokenImg(this);"></div><div class="list-item__center" onclick="loadPost(' + "'" + id + "'" + ')"><div class="list-item__title sinhala">' + title + '</div><div class="list-item__subtitle" style="margin-top:5px;">' + source + " - " + datetime + '</div></div></li>';
  return (html);
}

// time loops
function checkNewPosts() {
  if (!localStorage.getItem('rated')) {
    showRateDialog();
  }

  var keys = Object.keys(newsList).sort();
  var newestId = keys[keys.length - 1];

  $.post(api, {
      action: "news_check",
    }, null, 'json')
    .done(function(data) {
      if (data.id !== newestId) {
        getNewsList(newestId, "null", "check");
      }
    })
    .fail(function() {
      ons.notification.alert("Unable to check news!");
    });

}

// fix things
function fixEscapedHtml(text) {
  return text
    .replace("&amp;", "&")
    .replace("&lt;", "<")
    .replace("&gt;", ">")
    .replace("&quot;", '"')
    .replace("&#039;", "'")
    .replace("&amp;#039;", "'");
}

function fixElements() {
  // fix broken elements of page & remove useless ones
  $("#post iframe").width('100%');
  // $("#post iframe").height('auto');
  $("#post img").width('100%');
  $("#post img").height('auto');
  $('img').attr('onerror', 'fixBrokenImg(this);');

  $("#post a, #post p").each(function() {
    var val = $(this).attr('href');
    if (val == null) {
      val = $(this).text();
    }
    if (val == null) {
      val = "null";
    }
    if (val.indexOf("fivefilters") >= 0 || val.indexOf("Viewers") >= 0) {
      $(this).remove();
    }
  });
}

function fixBrokenImg(img) {
  // when image error happen, set default img
  img.src = "./img/sources/default.png";
  return true;
}

// handle events
document.addEventListener("offline", onOffline, false);

function onOffline() {
  if (Object.keys(newsList).length == 0) {
    ons.notification.alert("You are offline!. Please connect to the internet.").then(function() {
      exitApp();
    });
  } else {
    ons.notification.alert("You are offline!. Some assets will not load properly.");
  }
}

// navigate between pages
function showNewsList() {
  // go back to news list
  showMainToolbar();
  $('#post').hide();
  $('#news-list').fadeIn();
  currentPage = "news-list";
}

function showPost(postId, data) {
  // set element values on post
  var source, datetime, title, mainImg, content, link;
  source = newsList[postId].source;
  datetime = newsList[postId].datetime;
  title = fixEscapedHtml(newsList[postId].title);
  content = fixEscapedHtml(data.post);
  link = data.link;
  $('#post-source, #post-source-bottom, #toolbar-title').text(source);
  $('#post-title').text(title);
  $('#post-datetime').text(datetime);
  $('#post-content').html(content);
  $('#post-link').attr("href", link);

  // fix element issus on post
  fixElements();

  // hide news list & show post
  $('#news-list').hide();
  showPostToolbar();
  $('#post').fadeIn();

  // scroll to top of page
  $('.page__content').scrollTop(0);
}

function openSourceURL() {
  var url = newsPosts[currentPostId].link;
  window.open(url, '_blank');
}

// show hide basic elements
function showToast(msg) {
  $('#outputToastMsg').text(msg);
  outputToast.toggle();
}

function hideToast() {
  outputToast.hide();
}

function showPostToolbar() {
  $('#toolbar-menu-toggler').hide();
  $('#toolbar-back').fadeIn();
  $('#toolbar-share').fadeIn();
  $('#toolbar-web').fadeIn();
}

function showMainToolbar() {
  $('#toolbar-back').hide();
  $('#toolbar-web').hide();
  $('#toolbar-share').hide();
  $('#toolbar-menu-toggler').fadeIn();
  $('#toolbar-title').text("Ceylon News");
}

// non functional
function showNotice() {
  var msg = "The content of this app comes from publicly available feeds of news sites and they retain all copyrights.\n\nThus, this app is not to be held responsible for any of the content displayed.\n\nThe owners of these sites can exclude their feeds with or without reason from this app by sending an email to me.";

  ons.notification.confirm(msg)
    .then(function(index) {
      if (index === 1) {
        localStorage.setItem('showNotice', true);
      } else {
        exitApp();
      }
    });
}

function sharePost() {
  window.plugins.socialsharing.share(newsList[currentPostId].title, null, null, " - Readmore @ " + newsPosts[currentPostId].link);
}

function showRateDialog() {
  AppRate.preferences = {
    displayAppName: 'Ceylon News',
    promptAgainForEachNewVersion: false,
    simpleMode: true,
    usesUntilPrompt: 4,
    storeAppURL: {
      android: 'market://details?id=xyz.navinda.ceylonnews'
    },
    customLocale: {
      title: "Would you mind rating %@?",
      message: "It won’t take more than a minute and helps to promote my app. Thanks for your support!",
      cancelButtonLabel: "No, Thanks",
      laterButtonLabel: "Remind Me Later",
      rateButtonLabel: "Rate It Now",
      yesButtonLabel: "Yes!",
      noButtonLabel: "Not really",
      appRatePromptTitle: 'Do you like using %@?',
      feedbackPromptTitle: 'Mind giving us some feedback?',
    },
    callbacks: {
      handleNegativeFeedback: null,
      onRateDialogShow: function(callback) {
        callback(1);
      },
      onButtonClicked: function(buttonIndex) {
        if (buttonIndex == 3 || buttonIndex == 1) {
          localStorage.setItem('rated', true);
        }
      }
    }
  };

  AppRate.promptForRating();
}

// handle slide menu (code from onsen ui)
window.fn = {};

window.fn.open = function() {
  var menu = document.getElementById('menu');
  menu.open();
};

window.fn.load = function(page) {
  var content = document.getElementById('content');
  var menu = document.getElementById('menu');
  content.load(page)
    .then(menu.close.bind(menu));
};
