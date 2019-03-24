// global variables
var newsList = {};
var newsPosts = {};
var selectedSource = "-1";
var lang = "sn";
var currentPage = "news-list";
var settings = {};
var currentPostId = null;
// fix for auto load when menu is open
var newsListAutoLoad;

ons.ready(function () {
  init();
});

function init() {
  onsenInit();
  noticeShow();
  if (!langCheck()) {
    langSelectorShow();
  } else {
    settingsCheck();
    sourcesLoad();
    newsListLoad("-1", "-1", "normal");
    setInterval(newsUpdateCheck, 60000);
    newsListOnScrollInit();
    bModeDefaultsSet();
    updateCheck();
  }
}

function onsenInit() {
  ons.disableDeviceBackButtonHandler();
  onsenSlideBarInit();
}

// set onisen slider
function onsenSlideBarInit() {
  var menu = document.getElementById("menu");
  window.fn = {};
  window.fn.open = function () {
    menu.open();
  };
  window.fn.load = function (page) {
    var content = document.getElementById("content");
    var menu = document.getElementById("menu");
    content.load(page).then(menu.close.bind(menu));
  };
  // disable post auto load when menu is open
  menu.addEventListener("postopen", function () {
    newsListAutoLoad = settings.newsListAutoLoad ? false : null;
  });

  menu.addEventListener("postclose", function () {
    newsListAutoLoad = settings.newsListAutoLoad ? true : null;
  });
}

// check language is set
function langCheck() {
  if (!localStorage.getItem("lang")) {
    return false;
  } else {
    lang = localStorage.getItem("lang");
    return true;
  }
}

// show language selector modal
function langSelectorShow() {
  var modal = $("#modalLangSelect");
  modal.show();
}

// set selected language
function langSet(lang) {
  localStorage.setItem("lang", lang);
  var modal = $("#modalLangSelect");
  modal.hide();
  window.location = "index.html";
}

// get & load sources from online
function sourcesLoad() {
  requestSend(
    "get",
    {
      action: "sources_list",
      lang: lang
    },
    function (sources) {
      for (var i = 0; i < sources.length; i++) {
        $("#menuSources").append(
          "<ons-list-item tappable onclick=\"sourceLoad('" +
          sources[i].source +
          "');\">" +
          sources[i].source +
          "</ons-list-item>"
        );
      }
    }
  );
}

// load news list
function newsListLoad(postId, source, mode) {
  toastToggle("Loading news list...", null);
  requestSend(
    "get",
    {
      action: "news_list",
      post_id: postId,
      source: source,
      mode: mode,
      lang: lang
    },
    function (data) {
      if (!isNullOrEmpty(data)) {
        newsListAdd(data, mode);
        htmlElementsFix();
        $("#btnLoadMore").fadeIn();
      } else {
        $("#btnLoadMore").fadeOut();
      }
      toastToggle(null, null);
    }
  );
}

// append to news list
function newsListAdd(data, mode) {
  // mode 
  // check = when new post is present
  // normal = request for more posts normally
  // load = load from global newList object
  for (var i = 0; i < data.length; i++) {
    newsList[data[i].id] = data[i];
    if (mode == "normal") {
      $("#newsList").append(newsListItemGet(data[i]));
    } else if (mode == "load") {
      $("#newsList").prepend(newsListItemGet(data[i]));
    } else {
      $("#newsList").prepend(newsListItemGet(data[i]));
      notificationShow(data[i]);
    }
    imgLoadingShow(data[i].id, data[i].mainImg);
  }
  settingsApply();
}

// generate li element for the news list
function newsListItemGet(post) {
  var id, source, datetime, title, mainImg;
  id = post.id;
  source = post.source;
  datetime = post.datetime;
  title = escapedHtmlFix(post.title);
  mainImg = "./img/loading.gif";
  var html =
    '<li id="' +
    id +
    '" class="list-item"><div class="list-item__left"><img class="list-item__thumbnail" src="' +
    mainImg +
    '" alt="mainImg"  onerror="brokenImgFix($(this));"></div><div class="list-item__center" onclick="postLoad(\'' +
    id +
    '\')"><div class="list-item__title sinhala">' +
    title +
    '</div><div class="list-item__subtitle" style="margin-top:5px;">' +
    source +
    " - " +
    datetime +
    "</div></div></li>";
  return html;
}

// event listener to detect end of the news list
function newsListOnScrollInit() {
  $('.page__content').on('scroll', function () {
    var isBottom = ($(this).scrollTop() + $(this).innerHeight() + 100 >= $(this)[0].scrollHeight);
    if (isBottom && (currentPage == "news-list") && settings.newsListAutoLoad && newsListAutoLoad) {
      newsLoadMore();
    }
  });
}

// refresh data
function newsRefresh() {
  if (currentPage == "news-list") {
    $("#btnLoadMore").hide();
    $("#newsList").empty();
    toastToggle("Loading posts...", null);
    newsListLoad("-1", "-1", "normal");
  } else if (currentPage == "post") {
    postGetOnline(currentPostId);
  }
  settingsApply();
}

function newsUpdateCheck() {
  var keys = Object.keys(newsList);
  var newestId = keys[keys.length - 1];
  requestSend(
    "get",
    {
      action: "news_check",
      lang: lang
    },
    function (data) {
      if (data.id > newestId) {
        newsListLoad(newestId, "-1", "check");
        toastToggle("New posts are available!.", 4000);
      }
    }
  );
}

// load post
function postLoad(postId) {
  content.load("./views/newsPost.html").then(function () {
    if (postId in newsPosts) {
      postGetOffline(postId);
    } else {
      postGetOnline(postId);
    }
    currentPage = "post";
    currentPostId = postId;
    settingsApply();
  });
}

// load already viewd posts
function postGetOffline(postId) {
  postSet(postId, newsPosts[postId]);
}

// load posts from online
function postGetOnline(postId) {
  toastToggle("Loading post...", null);
  requestSend(
    "get",
    {
      action: "news_post",
      post_id: postId,
      lang: lang
    },
    function (data) {
      newsPosts[data.id] = data;
      postSet(postId, data);
      toastToggle(null, null);
    }
  );
}

// get news list from custom source
function sourceLoad(source) {
  selectedSource = source;
  toastToggle("Loading posts...", null);
  $("#btnLoadMore").hide();
  $("#newsList").empty();
  newsListLoad("-1", source, "normal");
  $("#toolbarTitle").text(source);
  menu.close();
}

// handle load more button
function newsLoadMore() {
  $("#btnLoadMore").hide();
  toastToggle("Loading more posts...", null);
  var keys = Object.keys(newsList);
  var oldestId = keys[0];
  newsListLoad(oldestId, selectedSource, "normal");
}

// show loading spinner while news list imgs load
function imgLoadingShow(id, img) {
  var tmpImg = new Image();
  var newsListImg = $("#" + id + " img");
  var imageLoaded = function () {
    $(newsListImg).attr("src", img);

  };
  var imageNotLoaded = function () {
    $(newsListImg).attr("src", "./img/sources/default.png");
  };
  tmpImg.onload = imageLoaded;
  tmpImg.onerror = imageNotLoaded;
  tmpImg.src = img;
}

// fix broken html tags when json parse
function escapedHtmlFix(text) {
  return text
    .replace("&amp;", "&")
    .replace("&lt;", "<")
    .replace("&gt;", ">")
    .replace("&quot;", '"')
    .replace("&#039;", "'")
    .replace("&amp;#039;", "'");
}

// fix useless elements in posts
function htmlElementsFix() {
  $("#post iframe").width("100%");
  $("#post img").width("100%");
  $("#post img").height("auto");
  $("#post img").attr("onerror", "brokenImgFix(this);");
  // remove useless elements
  elementRemover("#post a, #post p", ["fivefilters", "Viewers"]); elementRemover("#post a, #post p", ["fivefilters", "Viewers"]);
  // fix print logo issue
  $("img").each(function () {
    var src = $(this).attr("src");
    if (src.indexOf("print.png") > -1) {
      brokenImgFix($(this));
    }
  });

  // site specific fixes
  // gossip lanka blank ad spaces
  try {
    $(".adsbygoogle").remove();
  } catch (e) { }
}

// remove elements
function elementRemover(selectors, array) {
  for (var item in array) {
    remove(array[item]);
  }

  function remove(str) {
    $(selectors).each(function () {
      var val =
        $(this).attr("href") == null ? $(this).text() : $(this).attr("href");
      if (val.indexOf(str) > -1) {
        $(this).remove();
      }
    });
  }
}

// when image error happen, set default img
function brokenImgFix(img) {
  $(img).attr("src", "./img/sources/default.png");
  return true;
}

// show news list
function newsListShow() {
  content.load("./views/newsList.html").then(function () {
    if (!isNullOrEmpty(newsList)) {
      newsListAdd(Object.values(newsList), "load");
      $("#btnLoadMore").fadeIn();
      newsListOnScrollInit();
      // scroll to last position
      if (currentPostId !== null) {
        var id = "#" + currentPostId;
        $(".page__content").scrollTop(($(id).offset().top) - 80);
        currentPostId = null;
      }
    } else {
      newsListLoad("-1", "-1", "normal");
    }
    currentPage = "news-list";

  });
}

function postSet(postId, data) {
  // set element values on post
  var source, datetime, title, mainImg, content, link;
  source = newsList[postId].source;
  datetime = newsList[postId].datetime;
  title = escapedHtmlFix(newsList[postId].title);
  content = escapedHtmlFix(data.post);
  link = data.link;
  $("#toolbarTitle, #postSource").text(source);
  $("#postTitle").text(title);
  $("#postDateTime").text(datetime);
  $("#postBody").html(content);
  $("#postLink").attr("href", link);

  // fix element issus on post
  htmlElementsFix();

  currentPage = "post";
}

// open original url to article
function sourceUrlOpen() {
  var url = newsPosts[currentPostId].link;
  window.open(url, "_blank");
}

// settings 
function settingsShow() {
  content.load("./views/settings.html").then(function () {
    settingHandlersReg();
    settingsUIupdate();
  });
  var menu = document.getElementById("menu");
  menu.close();
  currentPage = "settings";
}

// show/hide toast in bottom
function toastToggle(msg, time) {
  var toast = $("#outputToast");
  var toastMsg = $("#outputToastMsg");

  if (isNullOrEmpty(msg) && isNullOrEmpty(time)) {
    toast.hide();
  } else if (!isNullOrEmpty(msg) && time !== null) {
    ons.notification.toast(msg, { timeout: time, animation: "ascend" });
  } else {
    toastMsg.text(msg);
    toast.show();
  }
}

function postShare() {
  window.plugins.socialsharing.share(
    newsList[currentPostId].title,
    null,
    null,
    " - Readmore @ " + newsPosts[currentPostId].link
  );
}

function notificationShow(post) {
  if (settings.notificationShow) {
    cordova.plugins.notification.local.schedule({
      title: post.title,
      text: post.source,
      foreground: true
    });

    cordova.plugins.notification.local.on("click", function () {
      postLoad(post.id);
    }, this);
  }
}

// background mode settings
function bModeDefaultsSet() {
  cordova.plugins.backgroundMode.setDefaults({
    title: "Ceylon News",
    text: "Running in the background",
  });
}

function requestSend(type, data, callback) {
  // type = request type
  // data = data to be send
  // callback = function to run after success
  if (settings.apiNew) {
    api = "http://35.211.9.240:3000/cn/v1.0";
  } else {
    api = "https://api.navinda.xyz/cn/v2.4/";
  }

  $.ajax({
    url: api,
    type: type,
    data: data,
    dataType: "json",
    timeout: 30000,
    success: function (data) {
      callback(data);
    },
    error: function () {
      toastToggle("Request failed!", 4000);
    }
  });
}

function isNullOrEmpty(input) {
  return jQuery.isEmptyObject(input);
}

// when user goes offline
document.addEventListener("offline", onOffline, false);
function onOffline() {
  if (isNullOrEmpty(newsList)) {
    ons.notification
      .alert("You are offline!. Please connect to the internet.")
      .then(function () {
        exitApp();
      });
  } else {
    ons.notification.alert(
      "You are offline!. Some assets will not load properly."
    );
  }
}

// disclamer notice
function noticeShow() {
  if (!localStorage.getItem("showNotice")) {
    var msg =
      "The content of this app comes from publicly available feeds of news sites and they retain all copyrights.\n\nThus, this app is not to be held responsible for any of the content displayed.\n\nThe owners of these sites can exclude their feeds with or without reason from this app by sending an email to me.";

    ons.notification.confirm(msg).then(function (index) {
      if (index === 1) {
        localStorage.setItem("showNotice", true);
      } else {
        exitApp();
      }
    });
  }
}
