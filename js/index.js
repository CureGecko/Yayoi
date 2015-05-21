/*
index.js
Yayoi

Created by Cure Gecko on 5/10/15.
Copyright 2015, Cure Gecko. All rights reserved.

Main logic for all pages.
*/

//The directory in which the site exists.
var SitePath = "/";
var APIPath = "/api/";

//Loads a page form API and saves the sate to the history.
function loadPageDetails(path, data, updateBrowser, addDataToURI) {
	$("#mainContainer").load(APIPath+path, data);
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

function updateUserInfo() {
	$("#userInfo").load(APIPath+"users", function(response, status, xhr) {
		var authenticated = $("#userInfo #authenticated").text()=="true";
		$("#loginMenu").css("display", (authenticated ? "none" : "block"));
		$("#logoutMenu").css("display", (!authenticated ? "none" : "block"));
		$("#uploadMenu").css("display", (!authenticated ? "none" : "block"));
	});
}

//After all the elements have been rendered.
$(document).ready(function() {
	//First load should have user info updated.
	updateUserInfo();

	//Load the first page from API.
	loadPageDetails(fullPath, firstLoadData, false, false);

	//Handle the login menu item to load the login page.
	$("#loadLogin").click(function(event) {
		loadPage("users/login");
		event.preventDefault();
	});

	//Handle the logout menu item to load the logout page.
	$("#loadLogout").click(function(event) {
		loadPage("users/logout");
		event.preventDefault();
	});

	//Handle the logout menu item to load the logout page.
	$("#loadUpload").click(function(event) {
		loadPage("uploads");
		event.preventDefault();
	});

	//Tag Auto Complete Listeners
	$("#tagAutoComplete").on("mouseenter", "li", function() {
		$("#tagAutoComplete li").removeClass("active");
		$(this).addClass("active");
	});

	$("#search_field").focus(function(event) {
		autoCompleteFor($(this), true);
	});
	$("#search_field").blur(function(event) {
		autoCompleteFor(null, true);
	});
});

//Tag Auto Complete
var autoCompleteField = null;
var autoCompleteWord = false;

function autoCompleteReplaceTag(tag) {
	$("#tagAutoComplete").hide();
	if (autoCompleteWord) {
		tag = tag.replace(/\s/g, "_");
		tag += " ";
		var text = autoCompleteField.val();
		var endRange = autoCompleteField[0].selectionStart;
		var textRange = text.substr(0, endRange);
		var startRange = textRange.lastIndexOf(" ")+1;
		var newText = text.substr(0, startRange)+tag+text.substr(endRange);
		autoCompleteField.val(newText);
		var newCursorPosition = endRange+(tag.length-(endRange-startRange));
		autoCompleteField[0].setSelectionRange(newCursorPosition, newCursorPosition);
	} else {
		autoCompleteField.val(tag);
	}
}

function autoCompleteReplace() {
	var tag = $("#tagAutoComplete .active");
	var value = tag.attr("value");
	var alias = tag.attr("alias");
	if (alias!="") {
		autoCompleteReplaceTag(alias);
	} else {
		autoCompleteReplaceTag(value);
	}
}

function autoCompleteKeydown(event) {
	var key = event.which;
	if (key==38) {//Up key
		event.preventDefault();
		var tag = $("#tagAutoComplete .active");
		var previousTag = tag.prev();
		if (previousTag.length!=0) {
			tag.removeClass("active");
			previousTag.addClass("active");
		}
	} else if (key==40) {//Down key
		event.preventDefault();
		var tag = $("#tagAutoComplete .active");
		var nextTag = tag.next();
		if (nextTag.length!=0) {
			tag.removeClass("active");
			nextTag.addClass("active");
		}
	} else if (key==9) {//Tab key
		event.preventDefault();
		autoCompleteReplace();
	}
}

function autoCompleteGetCompletionTag() {
	var text = autoCompleteField.val();
	if (autoCompleteWord) {
		text = text.substr(0, autoCompleteField[0].selectionStart);
		return text.substr(text.lastIndexOf(" ")+1);
	}
	return text;
}

function autoCompleteVisable() {
	return $("#tagAutoComplete").is(":visible");
}

function autoCompleteInput(event) {
	var tag = autoCompleteGetCompletionTag();
	if (tag=="") {
		$("#tagAutoComplete ul").html("");
		$("#tagAutoComplete").hide();
		return;
	}
	$("#tagAutoCompleteAPI").load(APIPath+"tags", {tag: tag}, function(response, status, xhr) {
		var autoCompleteList = $("#tagAutoComplete ul");
		autoCompleteList.html("");
		var tags = $("#tagAutoCompleteAPI #tags .tag");
		if (tags.length==0) {
			return;
		}
		for (var i=0; i<tags.length; i++) {
			var tagInfo = $(tags[i]);
			var id = tagInfo.find(".id").text();
			var value = tagInfo.find(".value").text();
			var alias = tagInfo.find(".alias").text();
			var tag = document.createElement("li");
			tag.setAttribute("tagID", id);
			tag.setAttribute("value", value);
			tag.setAttribute("alias", alias);
			if (i==0) {
				tag.setAttribute("class", "active");
			}
			if (autoCompleteWord) {
				tag.textContent = value.replace(/\s/g, "_");
			} else {
				tag.textContent = value;
			}
			autoCompleteList.append(tag);
		}
		$("#tagAutoComplete").show();
	});
}

function autoCompleteFor(field, word) {
	$("#tagAutoComplete").hide();
	if (autoCompleteField!=null) {
		autoCompleteField.unbind("keydown", autoCompleteKeydown);
		autoCompleteField.unbind("input paste", autoCompleteInput);
	}
	autoCompleteField = field;
	autoCompleteWord = word;
	if (autoCompleteField==null) {
		return
	}
	var offset = autoCompleteField.offset();
	var width = autoCompleteField.outerWidth(true);
	var height = autoCompleteField.outerHeight(true);
	$("#tagAutoComplete").css({top: offset.top+height, left: offset.left, width: width});

	autoCompleteField.keydown(autoCompleteKeydown);
	autoCompleteField.bind("input paste", autoCompleteInput);
}

//Special controls for post editing.
jQuery.fn.extend({
	initRating: function() {
		return this.each(function() {
		var input = $(this);
			input.hide();
			input.after("<div class=\"btn-group ratingButtons\" role=\"group\" aria-label=\"...\">"+
							"<button type=\"button\" class=\"btn btn-success\" value=\"s\">Safe</button>"+
							"<button type=\"button\" class=\"btn btn-info\" value=\"q\">Questionable</button>"+
							"<button type=\"button\" class=\"btn btn-warning\" value=\"e\">Explicit</button>"+
							"<button type=\"button\" class=\"btn btn-danger\" value=\"u\">Unrated</button>"+
						"</div>");
			$(this).parent().find(".ratingButtons button").click(function(event) {
				$(this).parent().find("button").removeClass("active");
				$(this).addClass("active");
				input.val($(this).val());
			});
	    });
	},
	initTags: function() {
		return this.each(function() {
			var input = $(this);
			input.hide();
			var inputContainer = $(document.createElement("div"));
			inputContainer.attr("class", "form-control tagsInputContainer");
			input.after(inputContainer);

			function newTagLine() {
				var newLine = document.createElement("div");
				newLine.setAttribute("class", "tagLine");
				var tagInput = document.createElement("input");
				tagInput.setAttribute("type", "text");
				tagInput.setAttribute("class", "tagInput");
				newLine.appendChild(tagInput);
				return newLine;
			}

			function parseTags() {
				inputContainer.html("");
				var tags = input.val().split(" ");
				if (tags.length==1 && tags[0]=="") {
					inputContainer.append(newTagLine());
				} else {
					for (var i=0; i<tags.length; i++) {
						var tag = tags[i];
						tag = tag.replace(/_/g, " ");
						if (tag!="") {
							var newLine = newTagLine();
							newLine.childNodes[0].setAttribute("value", tag);
							if (!/^[A-Za-z0-9]+[A-Za-z0-9\-\s\(\):!?]*$/i.test(tag)) {
								newLine.childNodes[0].setAttribute("class", "tagInput badTag");
							}
							inputContainer.append(newLine);
						}
					}
				}
			}
			parseTags();

			//Call .change() to parse any programmatic changes.
			input.change(function(event) {
				parseTags();
			});

			inputContainer.on("keydown", ".tagInput", function(event) {
				var key = event.which;
				//console.log(key);
				if (key==13) {//Enter key
					event.preventDefault();
					var line = $(this).parent();
					var newLine = newTagLine();
					line.after(newLine)
					newLine.childNodes[0].focus();
				} else if (key==38 && !autoCompleteVisable()) {//Up key
					event.preventDefault();
					var line = $(this).parent();
					var previousLine = line.prev()
					if (previousLine.length!=0) {
						previousLine.find("input").focus();
					}
				} else if (key==40 && !autoCompleteVisable()) {//Down key
					event.preventDefault();
					var line = $(this).parent();
					var nextLine = line.next();
					if (nextLine.length!=0) {
						nextLine.find("input").focus();
					}
				} else if (key==8 && $(this).val()=="") {//Delete key
					event.preventDefault();
					var line = $(this).parent();
					var previousLine = line.prev()
					var nextLine = line.next();
					if (previousLine.length!=0) {
						previousLine.find("input").focus();
						line.remove()
					} else if (nextLine.length!=0) {
						nextLine.find("input").focus();
						line.remove()
					}
				}
			});

			inputContainer.on("input paste", ".tagInput", function(event) {
				var tag = $(this).val();
				tag = tag.replace(/_/g, " ");
				$(this).val(tag);
				if (/^[A-Za-z0-9]+[A-Za-z0-9\-\s\(\):!?]*$/i.test(tag)) {
					$(this).removeClass("badTag");
				} else {
					$(this).addClass("badTag");
				}
				var tagInputs = inputContainer.find(".tagInput");
				var tags = "";
				for (var i=0; i<tagInputs.length; i++) {
					var tag = $(tagInputs[i]).val();
					tag = tag.replace(/\s/g, "_");
					if (i==0) {
						tags += tag;
					} else {
						tags += " "+tag;
					}
				}
				input.val(tags);
			});

			inputContainer.on("focus", ".tagInput", function(event) {
				autoCompleteFor($(this), false);
			});

			inputContainer.on("blur", ".tagInput", function(event) {
				autoCompleteFor(null, false);
			});
		});
	}
});