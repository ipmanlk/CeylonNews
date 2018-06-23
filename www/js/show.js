//global variables
var title, link;

//check if device is ready
document.addEventListener("deviceready", onDeviceReady, false);

//when device is ready
function onDeviceReady() {

    //disable back button and apply custom command
    document.addEventListener("backbutton", function (e) {
        e.preventDefault();
        showLoading(true);
        location.replace('index.html');
    }, false);

    //hide content
    $('#content').hide();

    //load user selected post
    loadPost();

    //show content
    $('#content').fadeIn();

}

//show user selected post
function loadPost() {
    //variables
    var id, newsData, post, dateTime, source;

    /* 
        get news psot item from url.
        format: http://show.html?id=1
    */
    id = ($(location).attr('href')).split('id=')[1];

    //get newsData from localStorage
    newsData = JSON.parse(localStorage.getItem('newsData'));

    //get details for the requested post id
    source = newsData[id]['source'];
    title = newsData[id]['title'];
    post = newsData[id]['post'];
    link = newsData[id]['link'];
    dateTime = newsData[id]['dateTime'];

    //if script tags contain in the post body, remove them
    var scriptRegEx = /<script[\s\S]*?>[\s\S]*?<\/script>/gi;
    if (post.search(scriptRegEx)) {
        post = post.replace(scriptRegEx, '');
    }

    //if ins tags contain in the post body, remove them
    var insRegEx = /<ins[^>]*>/g;
    if (post.search(insRegEx)) {
        post = post.replace(insRegEx, '');
    }

    //if useless text contain in the post body, remove them
    post = post.replace("Let's block ads!", "");
    post = post.replace("(Why?)", "");

    //load relevent data to html elements in the body
    $("#source").text(source);
    $("#title").text(title);
    $("#post").html(post);
    $('#dateTime').text(dateTime);
    $('#postSource').html('Source : <b>' + source + '</b> - ' + '<a href="' + link + '">Original Post</a>');
}

//navigation for menus
function navigate(action) {
    switch (action) {
        case "home":
            showLoading(true);
            location.replace('index.html');
            break;
        case "refresh":
            location.replace($(location).attr('href'));
            break;
        case "share":
            window.plugins.socialsharing.share(title + " - Read @", null, null, link);
            break;
    }
}