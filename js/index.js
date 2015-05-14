/*
index.js
Yayoi

Created by Cure Gecko on 5/10/15.
Copyright 2015, Cure Gecko. All rights reserved.

Main logic for all pages.
*/

var SitePath = "/";

$(document).ready(function() {
	var fullPath = window.location.pathname.replace("/", "");
	if (fullPath.length!=0 && fullPath.substr(fullPath.length-1,1)=="/") {
		fullPath = fullPath.substr(0,fullPath.length-1);
	}
	var path = fullPath.split("/");
	console.log(path);

	$("#loadLogin").click(function(event) {
		$("#mainContainer").load("/api/users/login");
		event.preventDefault();
	});
});