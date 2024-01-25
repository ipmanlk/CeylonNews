ons.ready(() => init());

const init = () => {
	setGlobalVars();
	loadSettings();
	showDisclaimer();
	if (!checkLang()) return;
	initOnsenComponents();
	loadNewsSources().then(() => {
		loadNewsList("online");
	});

	$("body").attr("style", `${data.lang}-font`);
};

const setGlobalVars = () => {
	window.vars = {
		currentPage: "news-list",
		newsList: {},
		newsPosts: {},
		currentPostId: null,
		selectedSourceName: null,
		loadMore: true,
		searchEnabled: true,
		cursor: undefined,
	};
};

const initOnsenComponents = () => {
	// use cordova back button handler instead
	ons.disableDeviceBackButtonHandler();

	// setup left side menu
	const menu = document.getElementById("sidemenu");

	// setup object for global functions
	window.fn = {};

	window.fn.openSideMenu = () => {
		menu.open();
	};

	window.fn.closeSideMenu = () => {
		menu.close();
	};

	window.fn.loadPage = (page) => {
		return new Promise((resolve, reject) => {
			const content = document.getElementById("page-content");
			content
				.load(page)
				.then(() => {
					menu.close.bind(menu);
					resolve();
				})
				.catch((e) => reject(e));
		});
	};

	// event listeners to handle news list loading when side menu is open
	menu.addEventListener("postopen", () => {
		vars.loadMore = false;
	});

	menu.addEventListener("postclose", () => {
		vars.loadMore = true;
	});

	document.addEventListener("backbutton", onBackKeyDown, false);

	document.addEventListener("offline", onOffline, false);
};

const loadNewsSources = () => {
	return new Promise((resolve, reject) => {
		if (!data.sources) {
			showOutputToast("Loading news sources....");
			sendRequest("/sources", { languages: data.lang })
				.then((sources) => {
					appendToSideMenuSources(
						sources.map((source) => ({ ...source, enabled: true }))
					);
					hideOutputToast();
					resolve();
				})
				.catch((e) => {
					reject("Unable to load news sources.");
				});
		} else {
			appendToSideMenuSources(data.sources);
			resolve();
		}
	});
};

const appendToSideMenuSources = (sources) => {
	sources.forEach((source) => {
		const isChecked = source.enabled ? "checked" : "";
		$("#ul-sidemenu-sources").append(`
				<ons-list-item tappable>
						<label class="left">
								<ons-checkbox id="chk-${source.name}" onchange="toggleSource('${source.name}')" ${isChecked}></ons-checkbox>
						</label>
						<span onclick="loadNewsFromSource('${source.name}')">${source.name}</span>
				</ons-list-item>
				`);

		if (source.enabled == undefined) source.enabled = true;
	});

	// save in data
	data.sources = sources;
	saveSettings();
};

const loadNewsList = (mode) => {
	if (mode == "online") {
		showOutputToast("Loading news list....");
		let sourcesStr = getSourcesStr();
		setLoadMore(false);

		// keyword for searching
		const keyword = $("#txtNewsSearch").val();

		sendRequest("/news", {
			sources: sourcesStr,
			keyword: keyword,
			languages: data.lang,
			cursor: window.vars.cursor,
		}).then(({ data: newsList, paging }) => {
			window.vars.cursor = paging.prev;

			if (newsList.length == 0 && keyword == "") {
				hideOutputToast();
				showTimedToast("Ooops!. Failed to find anything on that.", 3000);
				return;
			}
			// when no search results found
			if (newsList.length == 0 && keyword !== "") {
				$("#ul-news-list").html(`
								<li class="list-item">
										<h4>No results found.</h4>
								</li>
								`);
			}
			appendToNewsList(newsList);
			hideOutputToast();
			setLoadMore(true);
		});
	}

	if (mode == "offline") {
		// load news list from global vars
		appendToNewsList(Object.values(vars.newsList).reverse());
	}
};

const appendToNewsList = (newsList) => {
	newsList.forEach((news) => {
		const time = formatTime(news.createdAt);

		if (settings["st-news-list-card-ui"]) {
			$("#ul-news-list").append(`
						<ons-card class="news-list-card" id="${news.id}" onclick="loadNewsPost('${news.id}')">
								<img id="img${news.id}" src="" style="width: 100%">
						<div class="title news-list-card-title">
								${news.title}
						</div>
						<div class="content news-list-card-content">
								${news.sourceName} - ${time}
						</div>
						</ons-card>  
						`);
		} else {
			$("#ul-news-list").append(`
						<li id="${news.id}" class="list-item" onclick="loadNewsPost('${news.id}')">
								<div class="list-item__left">
										<img id="img${news.id}" class="list-item__thumbnail" src="./img/loading.gif">
								</div>
								<div class="list-item__center">
										<div class="list-item__title">
												${news.title}
										</div>
										<div class="list-item__subtitle" style="margin-top:5px;">
												${news.sourceName} - ${time}
										</div>
								</div>
						</li>
						`);
		}

		// load news list item thumbnail
		loadNewsListItemImg(news.id, news.thumbnailURL);

		// store in the global vars mapped by news ids
		vars.newsList[news.id] = news;

		applySettings();
	});

	initNewsListScrollListener();

	setLoadMore(true);
};

const loadNewsListItemImg = (newsId, mainImg) => {
	// create new image obj
	const tmpImg = new Image();
	// select relevant news list item image
	const newsListItemImg = $(`#img${newsId}`);

	// once image has been loaded to tempImg (cached)
	const imageLoaded = () => {
		$(newsListItemImg).attr("src", mainImg);
	};

	// if image download failed
	const imageNotLoaded = () => {
		$(newsListItemImg).attr("src", "./img/sources/default.png");
	};

	tmpImg.onload = imageLoaded;
	tmpImg.onerror = imageNotLoaded;

	// load main image to the tempImg
	tmpImg.src = mainImg;
};

const loadNewsPost = (newsId) => {
	if (vars.newsPosts[newsId]) {
		showNewsPost(newsId, vars.newsPosts[newsId]);
	} else {
	}
	sendRequest(`/news/${newsId}`).then((newsPost) => {
		showNewsPost(newsId, newsPost);
		vars.newsPosts[newsId] = newsPost;
	});
};

const showNewsPost = (newsId, newsPost) => {
	fn.loadPage("./views/newsPost.html").then(() => {
		$("#lbl-toolbar-title").text(vars.newsList[newsId].sourceName);
		$("#lbl-news-post-title").html(vars.newsList[newsId].title);
		$("#lbl-news-post-datetime").text(
			formatTime(vars.newsList[newsId].createdAt)
		);
		$("#lbl-news-post-body").html(newsPost.contentHTML);
		$("#lbl-news-post-source").text(vars.newsList[newsId].sourceName);

		// store in the global vars mapped by news ids;
		vars.currentPostId = newsId;

		// change current page
		vars.currentPage = "news-post";

		// apply settings
		applySettings();
	});
};

const loadNewsFromSource = (sourceName) => {
	// set global source id
	vars.selectedSourceName = sourceName;

	showOutputToast("Loading news list....");

	// disable searching for source
	$("#txtNewsSearch").val("");
	$("#txtNewsSearch").hide();

	sendRequest("/news", {
		sources: sourceName,
		keyword: "",
		languages: data.lang,
		cursor: window.vars.cursor,
	}).then(({ data: newsList, paging }) => {
		window.vars.cursor = paging.prev;
		// clear saved news list
		vars.newsList = {};
		$("#ul-news-list").empty();
		appendToNewsList(newsList);
		setLoadMore(true);
		hideOutputToast();
	});

	// Update title
	$("#lbl-toolbar-title").text(sourceName);

	// close side menu
	fn.closeSideMenu();
};

const loadMore = () => {
	// if loadmore is disabled (no news to load)
	if (!vars.loadMore) return;

	// fix duplicate news item show up (this is enabled on appendToNewsList)
	setLoadMore(false);

	showOutputToast("Loading news list....");

	const sourcesStr =
		vars.selectedSourceName == null ? getSourcesStr() : vars.selectedSourceName;

	const keyword = $("#txtNewsSearch").val();

	sendRequest("/news", {
		sources: sourcesStr,
		keyword: keyword,
		languages: data.lang,
		cursor: window.vars.cursor,
	}).then(({ data: newsList, paging }) => {
		window.vars.cursor = paging.prev;
		hideOutputToast();
		if (newsList.length == 0) {
			// if there aren't any more news items
			showTimedToast("You have reached the end :).", 1000);
			return;
		}
		appendToNewsList(newsList);
	});
};

const toggleSource = (sourceName) => {
	const source = data.sources.find((source) => source.name == sourceName);
	if (!source) {
		return;
	}

	data.sources.every((source) => {
		if (source.name == sourceName) {
			source.enabled = !source.enabled;
			return false;
		}
		return true;
	});

	saveSettings();
};

const getSourcesStr = () => {
	// this will generate a string for api calls
	let sourcesStr = "";
	data.sources.forEach((source) => {
		if (!source.enabled) return;
		sourcesStr += `${source.name},`;
	});

	// remove final comma
	sourcesStr = sourcesStr.substring(0, sourcesStr.length - 1);

	return sourcesStr;
};

const showNewsList = () => {
	fn.loadPage("./views/newsList.html").then(() => {
		loadNewsList("offline");
		// change current page
		vars.currentPage = "news-list";

		// show load more button
		if (vars.loadMore && !settings["st-news-list-autoload"]) {
			$("#btn-news-list-loadmore").fadeIn();
		}

		applySettings();

		initNewsListScrollListener();

		// scroll to last position
		if (vars.currentPostId !== null) {
			const id = `#${vars.currentPostId}`;
			$(".page__content").scrollTop($(id).offset().top - 80);
			vars.currentPostId = null;
		}

		// reset selected source
		vars.selectedSourceName = null;
	});
};

const searchNews = () => {
	if (!vars.searchEnabled) return;

	// request news list with keyword
	$("#ul-news-list").empty();
	loadNewsList("online");

	// desiable search
	vars.searchEnabled = false;

	// NOTE: Temporary solution to prevent duplicate search results & search requests
	// enable search after 1s
	setTimeout(() => {
		vars.searchEnabled = true;
	}, 1000);
};

const checkLang = () => {
	if (!data.lang) {
		showLangSelectModal();
		// if there is no language set, disable modal dismiss
		$("#btn-modal-about-close").hide();
		return false;
	}
	return true;
};

const selectLang = (lang) => {
	data.lang = lang;
	delete data.sources;
	loadNewsSources().then(() => {
		saveSettings();
		window.location = "./index.html";
	});
};

const showLangSelectModal = () => {
	const modal = $("#modal-langselect");
	modal.show();
};

const showOutputToast = (text) => {
	const toast = $("#toast-bottom");
	const toastText = $("#lbl-toast-bottom");
	toastText.text(text);
	toast.show();
};

const hideOutputToast = () => {
	const toast = $("#toast-bottom");
	toast.hide();
};

const showTimedToast = (text, ms) => {
	ons.notification.toast(text, { timeout: ms, animation: "ascend" });
};

const setLoadMore = (isEnabled) => {
	if (isEnabled && !settings["st-news-list-autoload"]) {
		$("#btn-news-list-loadmore").fadeIn();
	} else {
		$("#btn-news-list-loadmore").fadeOut();
	}

	vars.loadMore = isEnabled;
};

const refreshNewsList = () => {
	vars.newsList = {};
	$("#ul-news-list").empty();
	loadNewsList("online");

	// update title
	$("#lbl-toolbar-title").text("Ceylon News");
};

const refreshNewsPost = () => {
	const newsId = vars.currentPostId;
	loadNewsPost(newsId);
};

const shareNewsPost = () => {
	const newsId = vars.currentPostId;
	const url = vars.newsPosts[newsId].url;
	window.plugins.socialsharing.share(url, null, null);
};

const loadOriginalPost = () => {
	const newsId = vars.currentPostId;
	const url = vars.newsPosts[newsId].url;
	cordova.InAppBrowser.open(url, "_blank", "location=yes");
};

const sideMenuAction = (action) => {
	fn.closeSideMenu();
	switch (action) {
		case "news":
			showNewsList();
			$("#ul-news-list").empty();
			break;

		case "settings":
			showSettingsPage();
			break;

		case "about":
			$("#modal-about").show();
			break;

		case "contact":
			window.location = "mailto:io@navinda.xyz?Subject=CeylonNews";
			break;
	}
};

const initNewsListScrollListener = () => {
	$(".page__content").on("scroll", (e) => {
		if (!settings["st-news-list-autoload"]) return;
		const target = e.target;
		const isBottom =
			$(target).scrollTop() + $(target).innerHeight() + 10 >=
			$(target)[0].scrollHeight;

		if (isBottom && vars.currentPage == "news-list" && vars.loadMore) {
			loadMore();
		}
	});
};

const onOffline = () => {
	if (Object.keys(vars.newsList).length == 0) {
		ons.notification
			.alert("You are offline!. Please connect to the internet.")
			.then(() => {
				navigator.app.exitApp();
			});
	} else {
		ons.notification.alert(
			"You are offline!. Some assets will not load properly."
		);
	}
};

const formatTime = (time) => {
	const DateTime = luxon.DateTime;
	return DateTime.fromISO(time)
		.setZone("Asia/Colombo")
		.toFormat("yyyy-LL-dd hh:mm a");
};

const showDisclaimer = () => {
	if (data.disclaimerAccepted) return;

	const msg = `The content of this app comes from publicly available feeds of news sites and they retain all copyrights.\n\nThus, this app is not to be held responsible for any of the content displayed.\n\nThe owners of these sites can exclude their feeds with or without reason from this app by sending an email to me.\n\nIf you wish to continue, press "OK". Otherwise, press "CANCEL" to exit the app.`;

	if (window.confirm(msg)) {
		data.disclaimerAccepted = true;
		saveSettings();
	} else {
		exitApp();
	}
};
