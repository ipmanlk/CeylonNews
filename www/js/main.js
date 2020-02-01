ons.ready(() => init());

const init = () => {
    setGlobalVars();
    loadSettings();
    if (!checkLang()) return;
    initOnsenComponents();
    loadNewsSources().then(() => {
        loadNewsList("online");
    });
}

const setGlobalVars = () => {
    window.vars = {
        currentPage: "news-list",
        newsList: {},
        newsPosts: {},
        currentPostId: null,
        selectedSourceId: null,
        loadMore: true
    };
}

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
            content.load(page).then(() => {
                menu.close.bind(menu);
                resolve();
            }).catch(e => reject(e));
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
}

const loadNewsSources = () => {
    return new Promise((resolve, reject) => {
        if (!data.sources) {
            showOutputToast("Loading news sources....");
            sendRequest({ action: "news-sources", lang: data.lang }).then(sources => {
                appendToSideMenuSources(sources);
                hideOutputToast();
                resolve();
            }).catch(e => {
                reject("Unable to load news sources.");
            });
        } else {
            appendToSideMenuSources(data.sources);
            resolve();
        }
    });
}

const appendToSideMenuSources = (sources) => {
    sources.forEach(source => {
        const isChecked = source.enabled ? "checked" : "";
        $("#ul-sidemenu-sources").append(`
        <ons-list-item tappable>
            <label class="left">
                <ons-checkbox input-id="chk-${source.id}" onchange="toggleSource('${source.id}')" ${isChecked}></ons-checkbox>
            </label>
            <span onclick="loadNewsFromSource('${source.id}')">${source.name}</span>
        </ons-list-item>
        `);

        if (source.enabled == undefined) source.enabled = true;
    });

    // save in data
    data.sources = sources;
    saveSettings();
}

const loadNewsList = (mode) => {
    if (mode == "online") {
        showOutputToast("Loading news list....");
        let sourcesStr = getSourcesStr();
        setLoadMore(false);
        sendRequest({ action: "news-list", sources: sourcesStr }).then(newsList => {
            if (newsList.length == 0) {
                hideOutputToast();
                showTimedToast("Ooops!. Failed to find anything on that.", 3000);
                return;
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
}

const appendToNewsList = (newsList) => {

    newsList.forEach(news => {
        if (settings["st-news-list-card-ui"]) {
            $("#ul-news-list").append(`
            <ons-card class="news-list-card" id="${news.id}" onclick="loadNewsPost('${news.id}')">
                <img id="img${news.id}" src="./img/loading.gif" style="width: 100%">
            <div class="title news-list-card-title">
                ${news.title}
            </div>
            <div class="content news-list-card-content">
                ${news.source} - ${news.time}
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
                        ${news.source} - ${news.time}
                    </div>
                </div>
            </li>
            `);
        }

        // load news list item thumbnail 
        loadNewsListItemImg(news.id, news.main_img);

        // store in the global vars mapped by news ids
        vars.newsList[news.id] = news;

        applySettings();
    });

    initNewsListScrollListener();

    setLoadMore(true);
}

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
}

const loadNewsPost = (newsId) => {

    if (vars.newsPosts[newsId]) {
        showNewsPost(newsId, vars.newsPosts[newsId]);
    } else {
    }
    sendRequest({ action: "news-post", post_id: newsId }).then(newsPost => {
        showNewsPost(newsId, newsPost);
        vars.newsPosts[newsId] = newsPost;
    });
}

const showNewsPost = (newsId, newsPost) => {
    fn.loadPage("./views/newsPost.html").then(() => {
        $("#lbl-toolbar-title").text(vars.newsList[newsId].source);
        $("#lbl-news-post-title").text(vars.newsList[newsId].title);
        $("#lbl-news-post-datetime").text(vars.newsList[newsId].time);
        $("#lbl-news-post-body").html(newsPost.post);
        $("#lbl-news-post-source").text(vars.newsList[newsId].source);
        $("#lbl-news-post-source-link").text(vars.newsList[newsId].link);

        // store in the global vars mapped by news ids
        vars.currentPostId = newsId;

        // change current page
        vars.currentPage = "news-post";

        // apply settings
        applySettings();
    });
}

const loadNewsFromSource = (sourceId) => {
    // set global source id
    vars.selectedSourceId = sourceId;

    showOutputToast("Loading news list....");

    sendRequest({ action: "news-list", sources: sourceId }).then(newsList => {
        // clear saved news list
        vars.newsList = {};
        $("#ul-news-list").empty();
        appendToNewsList(newsList);
        setLoadMore(true);
        hideOutputToast();
    });

    // close side menu
    fn.closeSideMenu();
}

const loadMore = () => {
    const newsIds = Object.keys(vars.newsList);
    const sourcesStr = (vars.selectedSourceId == null) ? getSourcesStr() : vars.selectedSourceId;
    const oldestNewsId = newsIds[0];

    // if loadmore is disabled (no news to load)
    if (!vars.loadMore) return;

    // fix duplicate news item show up (this is enabled on appendToNewsList)
    setLoadMore(false);

    showOutputToast("Loading news list....");

    sendRequest({ action: "news-list-old", news_id: oldestNewsId, sources: sourcesStr }).then(newsList => {
        hideOutputToast();
        if (newsList.length == 0) {
            // if there aren't any more news items 
            showTimedToast("You have reached the end :).", 1000);
            return;
        }
        appendToNewsList(newsList);
    });
}

const toggleSource = (sourceId) => {
    let isEnabled;
    if ($(`#chk-${sourceId}`)[0].checked) {
        isEnabled = true;
    } else {
        isEnabled = false;
    }

    data.sources.every(source => {
        if (source.id == sourceId) {
            source.enabled = isEnabled;
            return false;
        }
        return true;
    });

    saveSettings();
}

const getSourcesStr = () => {
    // this will generate a string for api calls
    let sourcesStr = "";
    data.sources.forEach(source => {
        if (!source.enabled) return;
        sourcesStr += `${source.id},`;
    });

    // remove final comma
    sourcesStr = sourcesStr.substring(0, sourcesStr.length - 1);

    return sourcesStr;
}

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
            $(".page__content").scrollTop(($(id).offset().top) - 80);
            vars.currentPostId = null;
        }
    });
}

const checkLang = () => {
    if (!data.lang) {
        showLangSelectModal();
        return false;
    }
    return true;
}

const selectLang = (lang) => {
    data.lang = lang;
    delete data.sources;
    loadNewsSources().then(() => {
        saveSettings();
        window.location = "./index.html";
    });
}

const showLangSelectModal = () => {
    const modal = $("#modal-langselect");
    modal.show();
}


const showOutputToast = (text) => {
    const toast = $("#toast-bottom");
    const toastText = $("#lbl-toast-bottom");
    toastText.text(text);
    toast.show();
}

const hideOutputToast = () => {
    const toast = $("#toast-bottom");
    toast.hide();
}

const showTimedToast = (text, ms) => {
    ons.notification.toast(text, { timeout: ms, animation: "ascend" });
}

const setLoadMore = (isEnabled) => {
    if (isEnabled && !settings["st-news-list-autoload"]) {
        $("#btn-news-list-loadmore").fadeIn();
    } else {
        $("#btn-news-list-loadmore").fadeOut();
    }

    vars.loadMore = isEnabled;
}

const refreshNewsList = () => {
    vars.newsList = {};
    $("#ul-news-list").empty();
    loadNewsList("online");
}

const refreshNewsPost = () => {
    const newsId = vars.currentPostId;
    loadNewsPost(newsId);
}

const shareNewsPost = () => {
    const newsId = vars.currentPostId;
    const url = vars.newsPosts[newsId].link;
    window.plugins.socialsharing.share(
        vars.newsList[newsId].title,
        null,
        null,
        " - Readmore @ " + url
    );
}

const loadOriginalPost = () => {
    const newsId = vars.currentPostId;
    const url = vars.newsPosts[newsId].link;
    window.open(url, "_blank");
}

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
    }
}

const initNewsListScrollListener = () => {
    $('.page__content').on('scroll', (e) => {
        if (!settings["st-news-list-autoload"]) return;
        const target = e.target;
        const isBottom = ($(target).scrollTop() + $(target).innerHeight() + 10 >= $(target)[0].scrollHeight);

        if (isBottom && (vars.currentPage == "news-list") && vars.loadMore) {
            loadMore();
        }

    });
}

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
}

const sendRequest = (data = { word: "cat" }, method = "get") => {
    return new Promise((resolve, reject) => {
        $.ajax({
            url: "http://35.211.9.240:3001/v1.0",
            method: method,
            dataType: "json",
            data: data,
            timeout: 10000
        }).done((res) => {
            resolve(res);
        }).fail(() => {
            hideOutputToast();
            showTimedToast("Request failed!.", 3000);
            reject("Request failed.");
        })
    });
}