//set main menu
$(document).on("pagecreate", "#index", function (e) {
    var activePage = this;
    $("#sideMenu").one("panelbeforeopen", function () {
        var screen = $.mobile.getScreenHeight(),
            header = $(".ui-header", activePage).hasClass("ui-header-fixed") ? $(".ui-header", activePage).outerHeight() - 1 : $(".ui-header", activePage).outerHeight(),
            footer = $(".ui-footer", activePage).hasClass("ui-footer-fixed") ? $(".ui-footer", activePage).outerHeight() - 1 : $(".ui-footer", activePage).outerHeight(),
            panelheight = screen - header - footer;
        $('.ui-panel').css({
            'top': header,
            'min-height': panelheight
        });
    });
});

//check if news is already downloaded & stored
var loadOnline;

if (localStorage.getItem('loadOnline') === null) {
    loadOnline = true;
} else {
    loadOnline = false;
}

//check if device is ready
document.addEventListener("deviceready", onDeviceReady, false);

//when device is ready
function onDeviceReady() {

    //show disclaimer
    showNotice();

    //focus latest news bar
    $('#newsDataInfo').click();

    //loading spinner
    showLoading(true);

    if (loadOnline) {
        //if news is not already loaded from online
        loadNewsOnline();
    } else {
        //load from local storage
        loadNewsOffline();
    }

    //check if offline
    if (!isOnline()) {
        navigator.notification.alert("Error: You are offline. Some assets will not load properly.", null, "Error", "Ok");
    }
}

//load news from the internet
function loadNewsOnline() {

    if (isOnline()) {

        //hide news list
        $("#newsList").hide();

        //ajax request to online php file 
        //https://ceylon-news.navinda.xyz/newsData.php
        $.ajax({
            type: 'post',
            url: '',
            dataType: 'json',
            timeout: 500000, //50s

            success: function (response) {

                getOnlineNews(response);

                //hide loading spinner
                showLoading(false);

                //show news list
                $("#newsList").fadeIn();

            },

            // handle errors
            error: function (obj) {

                showLoading(false);

                // show error msg (native with cordova)
                var response = String(obj.responseText);

                if (response != "undefined") {
                    navigator.notification.alert(response, null, "Error", "Ok");
                } else {
                    navigator.notification.alert("Error: Server is unreachable.", null, "Error", "Ok");
                }

            }
        });

    } else {
        //offline msg
        navigator.notification.alert("Error: You are offline!. Please connect to the Internet.", null, "Error", "Ok");
    }
}

function getOnlineNews(response) {
    //object for offline news
    var newsData = {};

    //counter for ofline news object adding
    var id = 0;

    //variables
    var source, dateTime, title, post, link, img;

    //iterate trough json response from php
    for (item in response) {
        //fill variables with data
        source = response[item].source;
        dateTime = response[item].dateTime;
        title = response[item].title;
        post = response[item].post;
        link = response[item].link;

        //strip slashes 
        link = link.replace(new RegExp("\\\\", "g"), "");

        img = extractImg(post, source);

        //dynamically add list items to list view
        $("#newsList").append('<li id="' + item + '"><a href="#" class="sinhala" onclick="goToNewsPost(' + id + ');"><img src="' + img + '"><h2 class="full-text">' + title + '</h2><p> ' + source + ' - ' + dateTime + ' </p></a></li>').listview('refresh');

        //add items to offline news object
        newsData[id] = { 'dateTime': dateTime, 'source': source, 'title': title, 'post': post, 'link': link, 'img': img };

        //increment counter
        id++;
    }

    //store news data object in local storage
    localStorage.setItem('newsData', JSON.stringify(newsData));
    /* 
        store load online false on local storage
        when loading this index.html next time, 
        data will be loaded from the local storage
    */
    localStorage.setItem('loadOnline', false);

    //store updated date time in local storage
    updatedDateTime = (new Date()).toISOString();
    localStorage.setItem('updatedDateTime', updatedDateTime);

}

//load news from local storage
function loadNewsOffline() {

    //hide news list
    $("#newsList").hide();

    var newsData = JSON.parse(localStorage.getItem('newsData'));

    //iterate through newsData object and add items dynamically to listview
    for (item in newsData) {
        $("#newsList").append('<li id="' + item + '"><a href="#" class="sinhala" onclick="goToNewsPost(' + item + ');"><img src="' + newsData[item]['img'] + '"><h2 class="full-text">' + newsData[item]['title'] + '</h2><p> ' + newsData[item]['source'] + ' - ' + newsData[item]['dateTime'] + ' </p></a></li>').listview('refresh');
    }

    //get news updated date time
    var updatedDateTime = localStorage.getItem('updatedDateTime');

    //format with moment.js
    updatedDateTime = moment(updatedDateTime).format("YYYY-MM-DD hh:mm A");

    //set news data info
    $('#newsDataInfo').text('Latest News - (Updated : ' + updatedDateTime + ')');

    //show news list
    $("#newsList").fadeIn();

    showLoading(false);

    //scroll to last position
    var scrollPosition = parseFloat(localStorage.getItem('scrollPosition'));
    $.mobile.silentScroll(scrollPosition);
    localStorage.setItem('scrollPosition', (0));
}

//extract img src values from news article html strings
function extractImg(html, source) {
    var rex = /<img[^>]+src="([^">]+)/g;
    var img = rex.exec(html);
    if (img !== null) {
        return (img[1]);
    } else {
        switch (source) {
            case "Gossip Lanka":
                return ('./img/sources/gossiplanka.png');
                break;
            case "Lanka C News":
                return ('./img/sources/cnews.png');
                break;
            case "Lankadeepa":
                return ('./img/sources/lankadeepa.png');
                break;
            case "News First":
                return ('./img/sources/news1st.png');
                break;
            case "Hiru News":
                return ('./img/sources/hiru.png');
                break;
            case "Ada Derana":
                return ('./img/sources/derana.png');
                break;
            default:
                return ('./img/sources/default.png');
        }
    }
}

//load selected news post
function goToNewsPost(id) {
    //remember scrolled position
    localStorage.setItem('scrollPosition', ($('#' + id).offset().top) - 80);
    //load post
    showLoading(true);
    location.replace("show.html?id=" + id);

}

//refresh
function refresh() {
    localStorage.removeItem('loadOnline');
    location.replace('index.html');
}

//check online status
function isOnline() {

    var networkState = navigator.connection.type;

    if (networkState == 'none') {
        return false;
    } else {
        return true;
    }

}

//Disclaimer msg
function onConfirm(buttonIndex) {
    if (buttonIndex === 1) {
        localStorage.setItem('showNotice', false);
    } else {
        navigator.app.exitApp();
    }
}

function showNotice() {
    if (localStorage.getItem('showNotice') === null) {
        var msg = "The content of this app comes from publicly available feeds of news sites and they retain all copyrights.\n\nThus, this app is not to be held responsible for any of the content displayed.\n\nThe owners of these sites can exclude their feeds with or without reason from this app by sending an email to me.";
        navigator.notification.confirm(
            msg,
            onConfirm,
            'Disclaimer',
            ['I UNDERSTAND', 'EXIT']
        );
    }
}
