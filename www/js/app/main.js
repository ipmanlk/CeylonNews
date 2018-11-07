var sc = "FFF";
// store news list, posts temp
var newsList = {};
var newsPosts = {};
var newsSources = {};
// current location in app
var currentPage = "news-list";
// current reading post
var currentPostID;

ons.ready(function() {
  // check disclamer notice
  if (!localStorage.getItem('showNotice')) {
    showNotice();
  }
  // disable built in back button handler of onsen
  ons.disableDeviceBackButtonHandler();
  // get news list
  showToast("Loading posts...");
  getNewsList("null", "null","normal");
  // check for new articles
  setTimeout(checkNewPosts, 10000);
});

function getNewsList(postID, source_id, mode) {
  // get news list from server
  $.ajax({
    type: 'post',
    data: {
      code:sc,
      news_li:0,
      post_id:postID,
      source_id:source_id,
      mode:mode
    },
    url: "https://pk.navinda.xyz/api/ceylon_news/v2.1/a.php",
    dataType: 'json',
    timeout: 60000, //60s
    success: function (data) {
      if (data.length !== 0) {
        for (item in data) {
          newsList[(data[item].id)] = data[item];

          if (mode == "normal") {
            $('#news-list-content').append(getNewListItem(data[item]));
          }

          if (mode == "check") {
            $('#news-list-content').prepend(getNewListItem(data[item]));
          }
        }

        if (mode == "check") {
          showToast("New articles are available!");
          setTimeout(hideToast, 4000);
        }

        if (mode == "normal") {
          hideToast();
        }
      } else {
        $('#load-more-btn').hide();
      }
      // show load more button
      $('#load-more-btn').fadeIn();

      // get sources
      if (newsSources[0] == null) {
        getSources();
      }
    },
    error: function () {
      ons.notification.alert("Unable to read feeds!");
    }
  });
}

function goToNewsList() {
  // go back to news list
  showMainToolbar();
  $('#post').hide();
  $('#news-list').fadeIn();
  currentPage = "news-list";
}

function loadMoreNews() {
  // load more posts
  $('#load-more-btn').hide();
  showToast("Loading more posts...");
  var keys = Object.keys(newsList);
  var oldestID = keys[0];
  getNewsList(oldestID, "null", "normal");
}

function getNewListItem(post) {
  // generate li element for list
  var id,source,datetime,title,mainImg;
  id = post.id;
  source = post.source;
  datetime = post.datetime;
  title = revertEscapedHtml(post.title);
  mainImg = post.mainImg;
  var html = '<li id="' + id + '" class="list-item"><div class="list-item__left"><img class="list-item__thumbnail" src="' + mainImg + '" alt="mainImg"  onerror="imgError(this);"></div><div class="list-item__center" onclick="loadPost(' + "'" + id + "'" + ')"><div class="list-item__title sinhala">' + title + '</div><div class="list-item__subtitle" style="margin-top:5px;">' + source + " - " + datetime + '</div></div></li>'
  return(html);
}

function loadPost(postID) {
  if (postID in newsPosts) {
    loadPostOffline(postID);
  } else {
    loadPostOnline(postID);
  }
  currentPage = "post";
  currentPostID = postID;
}

function loadPostOffline(postID) {
  showPost(postID, newsPosts[postID]);
}

function loadPostOnline(postID) {
  // get full post from server
  showToast("Loading post...");
  $.ajax({
    type: 'post',
    data: {
      news_pst:0,
      code:sc,
      post_id:postID
    },
    url: "https://pk.navinda.xyz/api/ceylon_news/v2.1/a.php",
    dataType: 'json',
    timeout: 60000, //60s
    success: function (data) {
      newsPosts[data.id] = data;
      showPost(postID, data);
      hideToast();
    }
  });

}

function showPost(postID, data) {
  // set element values on post
  var source,datetime,title,mainImg, content, link;
  source = newsList[postID].source;
  datetime = newsList[postID].datetime;
  title = revertEscapedHtml(newsList[postID].title);
  content = revertEscapedHtml(data.post);
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

function getSources() {
  // get sources and append them to menu
  $.post("https://pk.navinda.xyz/api/ceylon_news/v2.1/a.php",
  {
    code: sc,
    news_s: 0
  },
  function(data){
    var JSONdata = JSON.parse(data);
    for (item in JSONdata) {
      $('#menu-sources').append('<ons-list-item tappable onclick="loadSource(\'' + JSONdata[item].id + '\');">' + JSONdata[item].source + '</ons-list-item>');
    }
  });
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

function imgError(image) {
  // when image error happen, set default img
  image.src = "./img/sources/default.png";
  return true;
}

function showToast(msg, action) {
  $('#outputToastMsg').text(msg);
  outputToast.toggle();
}

function hideToast() {
  outputToast.hide();
}

function fixElements() {
  // fix broken elements of page & remove useless ones
  $("#post iframe").width('100%');
  // $("#post iframe").height('auto');
  $("#post img").width('100%');
  $("#post img").height('auto');
  $('img').attr('onerror', 'imgError(this);');

  $("#post a, #post p").each(function(){
    var val = $(this).attr('href');
    if (val == null) {val=$(this).text()}
    if (val == null) {val="null"}
    if (val.indexOf("fivefilters") >= 0 || val.indexOf("Viewers") >= 0) {
      $(this).remove();
    }
  });
}

function checkNewPosts() {
  var keys = Object.keys(newsList);
  var newestID = keys[keys.length - 1];
  getNewsList(newestID, "null", "check");
  setTimeout(checkNewPosts, 900000);
}

function sharePost() {
  window.plugins.socialsharing.share(newsList[currentPostID].title, null, null, " - Readmore @ " + newsPosts[currentPostID].link);
}

function openSourceURL() {
  var url = newsPosts[currentPostID].link;
  window.open(url, '_blank');
}

function revertEscapedHtml(text) {
  return text
  .replace("&amp;", "&")
  .replace("&lt;", "<" )
  .replace("&gt;", ">")
  .replace("&quot;", '"')
  .replace("&#039;", "'");
}

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

