/*
index.js
Yayoi

Created by Cure Gecko on 5/10/15.
Copyright 2015, Cure Gecko. All rights reserved.

Main logic for all pages.
*/

$(document).ready(function() {
	$("#loadLogin").click(function(event) {
		$("#mainContainer").load("/api/users/login");
		event.preventDefault();
	});
});