/*******************************************************************************
 * Copyright (c) 2008, 2021 SAP AG and IBM Corporation.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License 2.0
 * which accompanies this distribution, and is available at
 * https://www.eclipse.org/legal/epl-2.0/
 *
 * SPDX-License-Identifier: EPL-2.0
 *
 * Contributors:
 *     SAP AG - initial API and implementation
 *******************************************************************************/
function hide(obj, a)
{
	var imageBase = document.getElementById('imageBase');
	var div = document.getElementById(a).style;
	
	if (div.display == "none")
	{
		div.display = "block";
		obj.firstChild.src = imageBase.value + 'opened.gif'
	}
	else
	{
		div.display = "none";
		obj.firstChild.src = imageBase.value + 'closed.gif'
	}
	obj.title = swapTitle(obj.title)
	obj.firstChild.alt = swapTitle(obj.firstChild.alt)
}

function swapTitle(title)
{
	if (title != null)
	{
		sep = title.indexOf(" / ")
		if (sep >= 0)
			title = title.substring(sep + 3) + title.substring(sep, sep + 3) + title.substring(0, sep);
	}
	return title
}

function preparepage()
{
	var W3CDOM = document.getElementById && document.createTextNode
	if (!W3CDOM)
		return;

	rendertrees();
	stripetables();
	collapsible();
	roles();
}

function rendertrees()
{
	var imageBase = document.getElementById('imageBase');

	var tables=document.getElementsByTagName('table');
	for (var i=0;i<tables.length;i++)
	{
		if(tables[i].className!='result')
			continue;
			
		var tbodies = tables[i].getElementsByTagName("tbody");
		for (var h = 0; h < tbodies.length; h++)
		{
			if (tbodies[h].className!='tree')
				continue;

			var trs = tbodies[h].getElementsByTagName("tr");
			for (var ii = 0; ii < trs.length; ii++)
			{
				treerow(imageBase.value, trs[ii]);
			}
		}
	}
}

function treerow(imageBaseValue, element)
{
	var cell = element.firstChild;
	var celltext = cell.firstChild;

	var code = celltext.data;
	
	if (typeof code=='undefined')
		return;
	
	for(var ii=0; ii<code.length; ii++)
	{
		var c = code.charAt(ii);
		
		var replace = document.createElement('img');
		replace.alt = c;
		replace.className = 'line'

	
		switch(c)
		{
		case "+":
			replace.src = imageBaseValue + "fork.gif";
			break;
		case ".":
			replace.src = imageBaseValue + "empty.gif";
			break;
		case "\\":
			replace.src = imageBaseValue + "corner.gif";
			break;
		case "|":
			replace.src = imageBaseValue + "line.gif";
			break;
		}
		cell.insertBefore(replace, celltext);
	}
	
	cell.removeChild(celltext);
}

function stripetables()
{
	var tables=document.getElementsByTagName('table');
	for (var i=0;i<tables.length;i++)
	{
		if(tables[i].className=='result')
		{
			stripe(tables[i]);
		}
	}
}

function hasClass(obj)
{
	var result = false;
	if (obj.getAttributeNode("class") != null)
	{
		result = obj.getAttributeNode("class").value;
	}
	return result;
}   

function stripe(table)
{
	var even = false;

	var tbodies = table.getElementsByTagName("tbody");
	for (var h = 0; h < tbodies.length; h++)
	{
		var trs = tbodies[h].getElementsByTagName("tr");
		for (var i = 0; i < trs.length; i++)
		{

			if (!hasClass(trs[i]))
			{
				trs[i].className = even ? "evenrow" : "oddrow";
			}
			even =  ! even;
		}
	}
}

function roles()
{
	// Not valid HTML4, but screen reader might benefit
	var element = document.getElementById("header");
	element.setAttribute("role","navigation");
	element = document.getElementById("content");
	element.setAttribute("role","main");
	element = document.getElementById("footer");
	element.setAttribute("role","contentinfo");
}

//
// collapsible list items
//

function collapsible()
{
	var imageBase = document.getElementById('imageBase');
	closedImage = imageBase.value + 'closed.gif';
	openedImage = imageBase.value + 'opened.gif';
	openedImageAlt = imageBase.title;
	closedImageAlt = swapTitle(openedImageAlt);
	nochildrenImage = imageBase.value + 'nochildren.gif';
	
	var uls = document.getElementsByTagName('ul');
	for (var i=0;i<uls.length;i++)
	{
		if(uls[i].className == 'collapsible_opened')
		{
			makeCollapsible(uls[i], 'block', 'collapsibleOpened', openedImage, openedImageAlt);
		}
		else if(uls[i].className == 'collapsible_closed')
		{
			makeCollapsible(uls[i], 'none', 'collapsibleClosed', closedImage, closedImageAlt);
		}
	}
}

function makeCollapsible(listElement, defaultState, defaultClass, defaultImage, defaultImageAlt)
{
	listElement.style.listStyle = 'none';

	var child = listElement.firstChild;
	while (child != null)
	{
		if (child.nodeType == 1)
		{
			var list = new Array();
			var grandchild = child.firstChild;
			while (grandchild != null)
			{
				if (grandchild.tagName == 'OL' || grandchild.tagName == 'UL')
				{
					grandchild.style.display = defaultState;
					list.push(grandchild);
				}
				grandchild = grandchild.nextSibling;
			}
			
			var node = document.createElement('img');

			if (list.length == 0)
			{
				node.setAttribute('src', nochildrenImage);
				node.setAttribute('alt', '');
			}
			else
			{
				node.setAttribute('src', defaultImage);
				/* No need to set the image text as set on the a */
				if (false && defaultImageAlt != null)
					node.setAttribute('alt', defaultImageAlt);
				else
					node.setAttribute('alt', '');
				var anode = document.createElement('a');
				anode.href = "#"
				anode.setAttribute('class', defaultClass);
				anode.onclick = createToggleFunction(anode,list);
				/* Set the img alt text and the a title */
				if (defaultImageAlt != null)
					anode.title = defaultImageAlt 
				anode.appendChild(node)
				node = anode
			}

			child.insertBefore(node,child.firstChild);
		}

		child = child.nextSibling;
	}
}

function createToggleFunction(toggleElement, sublistElements)
{
	return function()
	{
		if (toggleElement.getAttribute('class')=='collapsibleClosed')
		{
			toggleElement.setAttribute('class','collapsibleOpened');
			toggleElement.firstChild.setAttribute('src',openedImage);
		}
		else
		{
			toggleElement.setAttribute('class','collapsibleClosed');
			toggleElement.firstChild.setAttribute('src',closedImage);
		}
		toggleElement.setAttribute('title', swapTitle(toggleElement.getAttribute('title')))
		toggleElement.firstChild.setAttribute('alt', swapTitle(toggleElement.firstChild.getAttribute('alt')))

		for (var i=0;i<sublistElements.length;i++)
		{
			sublistElements[i].style.display = (sublistElements[i].style.display=='block') ? 'none' : 'block';
		}
		return false
	}
}
