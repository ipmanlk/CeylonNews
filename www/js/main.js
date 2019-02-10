// global variables
var newsList = {};
var newsPosts = {};
var selectedSource = "-1";
var lang = "-1";
var currentPage = "news-list";

ons.ready(function () {
	init();
});


function init() {
	onsenInit();
	noticeShow();
	if (!langCheck()) {
		langSelectorShow();
	} else {
		sourcesLoad();
		newsListLoad("-1", "-1", "normal");
		appCoverImgLoad();
		setInterval(newsUpdateCheck, 60000);
	}
}

function onsenInit() {
	ons.disableDeviceBackButtonHandler();
	onsenSlideBarInit();
}

// set onisen slider
function onsenSlideBarInit() {
	window.fn = {};
	window.fn.open = function () {
		var menu = document.getElementById('menu');
		menu.open();
	};
	window.fn.load = function (page) {
		var content = document.getElementById('content');
		var menu = document.getElementById('menu');
		content.load(page)
			.then(menu.close.bind(menu));
	};
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
	var modal = $('#modalLangSelect');
	modal.show();
}

// set selected language
function langSet(lang) {
	localStorage.setItem("lang", lang);
	var modal = $('#modalLangSelect');
	modal.hide();
	window.location = "./main.html";
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
				$('#menu-sources').append('<ons-list-item tappable onclick="sourceLoad(\'' + sources[i].source + '\');">' + sources[i].source + '</ons-list-item>');
			}
		}
	);
}

//  load cover img
function appCoverImgLoad() {
	requestSend(
		"get",
		{
			action: "cover_img"
		},
		function (data) {
			$('#coverImg').attr('src', data.img);
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
				$('#load-more-btn').fadeIn();
			} else {
				$('#load-more-btn').fadeOut();
			}
			toastToggle(null, null);
		}
	);
}

// append to news list
function newsListAdd(data, mode) {
	for (var i = 0; i < data.length; i++) {
		newsList[(data[i].id)] = data[i];
		if (mode == "normal") {
			$('#news-list-content').append(newsListItemGet(data[i]));
		} else {
			$('#news-list-content').prepend(newsListItemGet(data[i]));
		}
	}
}

// generate li element for the news list
function newsListItemGet(post) {
	var id, source, datetime, title, mainImg;
	id = post.id;
	source = post.source;
	datetime = post.datetime;
	title = escapedHtmlFix(post.title);
	mainImg = post.mainImg;
	var html = '<li id="' + id + '" class="list-item"><div class="list-item__left"><img class="list-item__thumbnail" src="' + mainImg + '" alt="mainImg"  onerror="brokenImgFix($(this));"></div><div class="list-item__center" onclick="postLoad(\'' + id + '\')"><div class="list-item__title sinhala">' + title + '</div><div class="list-item__subtitle" style="margin-top:5px;">' + source + " - " + datetime + '</div></div></li>';
	return (html);
}

// refresh data
function newsRefresh() {
	if (currentPage == "news-list") {
		$('#load-more-btn').hide();
		$('#news-list-content').empty();
		toastToggle("Loading posts...", null);
		newsListLoad("-1", "-1", "normal");
	} else if (currentPage == "post") {
		postGetOnline(currentPostId);
	}
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
	if (postId in newsPosts) {
		postGetOffline(postId);
	} else {
		postGetOnline(postId);
	}
	currentPage = "post";
	currentPostId = postId;
}

// load already viewd posts
function postGetOffline(postId) {
	postShow(postId, newsPosts[postId]);
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
			postShow(postId, data);
			toastToggle(null, null);
		}
	);
}

// get news list from custom source
function sourceLoad(source) {
	selectedSource = source;
	toastToggle("Loading posts...", null);
	$('#load-more-btn').hide();
	$('#news-list-content').empty();
	newsListLoad("-1", source, "normal");
	$('#toolbar-title').text(source);
	menu.close();
}

// handle load more button
function newsLoadMore() {
	$('#load-more-btn').hide();
	toastToggle("Loading more posts...", null);
	var keys = Object.keys(newsList);
	var oldestId = keys[0];
	newsListLoad(oldestId, selectedSource, "normal");
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
	$("#post iframe").width('100%');
	$("#post img").width('100%');
	$("#post img").height('auto');
	$('img').attr('onerror', 'brokenImgFix(this);');
	// remove useless elements
	elementRemover('#post a, #post p', ["fivefilters", "Viewers"]);
	// fix print logo issue
	$("img").each(function () {
		var src = $(this).attr('src');
		if (src.indexOf("print.png") > -1) {
			brokenImgFix($(this));
		}
	});

	// site specific fixes
	// gossip lanka blank ad spaces
	try {
		$('.adsbygoogle').remove();
	} catch (e) { }
}

// remove elements
function elementRemover(selectors, array) {
	for (var item in array) {
		remove(array[item]);
	}

	function remove(str) {
		$(selectors).each(function () {
			var val = $(this).attr('href') == null ? $(this).text() : $(this).attr('href');
			if (val.indexOf(str) > -1) {
				$(this).remove();
			}
		});
	}
}

// when image error happen, set default img
function brokenImgFix(img) {
	$(img).attr('src', "./img/sources/default.png");
	return true;
}

// show news list
function newsListShow() {
	mainToolbarShow();
	$('#post').hide();
	$('#news-list').fadeIn();
	currentPage = "news-list";
}

function postShow(postId, data) {
	// set element values on post
	var source, datetime, title, mainImg, content, link;
	source = newsList[postId].source;
	datetime = newsList[postId].datetime;
	title = escapedHtmlFix(newsList[postId].title);
	content = escapedHtmlFix(data.post);
	link = data.link;
	$('#post-source, #post-source-bottom, #toolbar-title').text(source);
	$('#post-title').text(title);
	$('#post-datetime').text(datetime);
	$('#post-content').html(content);
	$('#post-link').attr("href", link);

	// fix element issus on post
	htmlElementsFix();

	// hide news list & show post
	$('#news-list').hide();
	postToolbarShow();
	$('#post').fadeIn();

	// scroll to top of page
	$('.page__content').scrollTop(0);
}

// open original url to article
function sourceUrlOpen() {
	var url = newsPosts[currentPostId].link;
	window.open(url, '_blank');
}

// show/hide toast in bottom
function toastToggle(msg, time) {
	var toast = $('#outputToast');
	var toastMsg = $('#outputToastMsg');

	if (isNullOrEmpty(msg) && isNullOrEmpty(time)) {
		toast.hide();
	} else if (!isNullOrEmpty(msg) && time !== null) {
		ons.notification.toast(msg, { timeout: time, animation: 'ascend' });
	} else {
		toastMsg.text(msg);
		toast.show();
	}
}


function postToolbarShow() {
	$('#toolbar-menu-toggler').hide();
	$('#toolbar-lang').hide();
	$('#toolbar-back').fadeIn();
	$('#toolbar-share').fadeIn();
	$('#toolbar-web').fadeIn();
}

function mainToolbarShow() {
	$('#toolbar-back').hide();
	$('#toolbar-web').hide();
	$('#toolbar-share').hide();
	$('#toolbar-menu-toggler').fadeIn();
	$('#toolbar-lang').fadeIn();
	$('#toolbar-title').text("Ceylon News");
}

function postShare() {
	window.plugins.socialsharing.share(newsList[currentPostId].title, null, null, " - Readmore @ " + newsPosts[currentPostId].link);
}

function requestSend(type, data, callback) {
	// type = request type
	// data = data to be send
	// callback = function to run after success 
	var api = "https://api.navinda.xyz/cn/v2.4/";
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
		ons.notification.alert("You are offline!. Please connect to the internet.").then(function () {
			exitApp();
		});
	} else {
		ons.notification.alert("You are offline!. Some assets will not load properly.");
	}
}


// disclamer notice
function noticeShow() {
	if (!localStorage.getItem('showNotice')) {
		var msg = "The content of this app comes from publicly available feeds of news sites and they retain all copyrights.\n\nThus, this app is not to be held responsible for any of the content displayed.\n\nThe owners of these sites can exclude their feeds with or without reason from this app by sending an email to me.";

		ons.notification.confirm(msg)
			.then(function (index) {
				if (index === 1) {
					localStorage.setItem('showNotice', true);
				} else {
					exitApp();
				}
			});
	}
}
