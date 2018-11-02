// store news list temp
var newsList = {}

$(document).ready(function() {
  getNewsList("null", "null");
});

function getNewsList(postID, source_id) {
  $.ajax({
    type: 'post',
    data: {
      code:"4a2204811369",
      post_id:postID,
      source_id:source_id
    },
    url: "https://pk.navinda.xyz/api/ceylon_news/v2.0/getNewsList.php",
    dataType: 'json',
    timeout: 60000, //60s
    success: function (data) {
      for (item in data) {
        $('#newsList').append(getNewListItem(data[item]));
        newsList[(data[item].id)] = data[item];
      }
      toast(null, "hide");
      // scroll to last position
      if (localStorage.getItem('scrollPosition') !== null) {
        var scrollPosition = parseFloat(localStorage.getItem('scrollPosition'));
        $('.page__content').scrollTop(scrollPosition);
      }
    }
  });
}

function goToNewsList() {
  fn.load('newsList.html');
  getNewsList('null', 'null');
}

function loadMoreNews() {
  toast("Loading more posts...", "show");
  var keys = Object.keys(newsList);
  var oldestID = keys[0];
  getNewsList(oldestID, "null");
}

function getNewListItem(post) {
  var id,source,datetime,title,mainImg;
  id = post.id;
  source = post.source;
  datetime = post.datetime;
  title = post.title;
  mainImg = post.mainImg;
  var html = '<li id="' + id + '" class="list-item"><div class="list-item__left"><img class="list-item__thumbnail" src="' + mainImg + '" alt="mainImg"  onerror="imgError(this);"></div><div class="list-item__center" onclick="loadPost(' + "'" + id + "'" + ')"><div class="list-item__title sinhala">' + title + '</div><div class="list-item__subtitle" style="margin-top:5px;">' + source + " - " + datetime + '</div></div></li>'
  return(html);
}

function loadPost(postID) {
  toast("Loading post...", "show");
  fn.load('post.html');
  $.ajax({
    type: 'post',
    data: {
      code:"4a2204811369",
      post_id:postID
    },
    url: "https://pk.navinda.xyz/api/ceylon_news/v2.0/getNewsPost.php",
    dataType: 'json',
    timeout: 60000, //60s
    success: function (data) {
      showPost(postID, data);
      toast(null, "hide");
    }
  });
  // save last scroll position
  localStorage.setItem('scrollPosition', ($('#' + postID).offset().top) - 70);
}

function showPost(postID, data) {
  var source,datetime,title,mainImg, content, link;
  source = newsList[postID].source;
  datetime = newsList[postID].datetime;
  title = newsList[postID].title;
  content = data.post;
  link = data.link;
  $('#post-source, #post-source-bottom').text(source);
  $('#post-title').text(title);
  $('#post-datetime').text(datetime);
  $('#post-content').html(content);
  $('#post-link').attr("href", link);
}

function imgError(image) {
  image.src = "https://image.freepik.com/free-icon/news-logo_318-38132.jpg";
  return true;
}

function toast(msg, action) {
  if (action == "show") {
    $('#outputToastMsg').text(msg);
    outputToast.toggle();
  } else {
    outputToast.hide();
  }
}

// handle slide menu
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

