/*
index.js
Yayoi

Created by Cure Gecko on 5/10/15.
Copyright 2015, Cure Gecko. All rights reserved.

Main logic for all pages.
*/

//The directory in which the site exists.
var SitePath = "/";

//Loads a page form API and saves the sate to the history.
function loadPageDetails(path, data, updateBrowser, addDataToURI) {
	$("#mainContainer").load(SitePath+"api/"+path, data);
	if (updateBrowser) {
		if (addDataToURI && data!=null && data!="") {
			var uriData = data;
			if (typeof data != "string") {
				var newData = "";
				for (key in data) {
					if (newData!="") {
						newData += "&";
					}
					newData += encodeURIComponent(key)+"="+encodeURIComponent(data[key]);
				}
				uriData = newData;
			}
			window.history.pushState({path: path, data: data}, "Yayoi", "/"+path+"?"+data);
		} else {
			window.history.pushState({path: path, data: data}, "Yayoi", "/"+path);
		}
	}
}
//Loads a page with path and data only.
function loadPageData(path, data) {
	loadPageDetails(path, data, true, true);
}
//Loads a page with path only.
function loadPage(path) {
	loadPageDetails(path, null, true, false)
}

//Parse the path requested.
var fullPath = window.location.pathname.replace(SitePath, "");
if (fullPath.length!=0 && fullPath.substr(fullPath.length-1,1)=="/") {
	fullPath = fullPath.substr(0,fullPath.length-1);
}
var path = fullPath.split("/").filter(function(s){ return s!="" });
//The request data that was sent on the first request. This is saved so that we do not have to push a new state for the first load.
var firstLoadData = window.location.search.replace("?", "");

//Handle loading pages from the history.
window.onpopstate = function(event) {
	if (event.state!=undefined) {
		if (event.state.path!=undefined) {
			loadPageDetails(event.state.path, event.state.data, false, false);
		}
	} else {
		loadPageDetails(fullPath, firstLoadData, false, false);
	}
};

//After all the elements have been rendered.
$(document).ready(function() {
	//Load the first page from API.
	loadPageDetails(fullPath, firstLoadData, false, false);

	//Handle the login menu item to load the login page.
	$("#loadLogin").click(function(event) {
		loadPage("users/login");
		event.preventDefault();
	});
});
