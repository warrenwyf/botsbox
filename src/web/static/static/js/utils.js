var utils = {};

utils.formatTimeInSecs = function(secs) {
	var today = moment().startOf('day');
	var time = moment(secs * 1000);

	str = '';

	if (!(today.year() == time.year() && today.month() == time.month() && today.date() == time.date())) {
		str = time.format('YYYY-MM-DD ');
	}

	str += time.format('HH:mm:ss');

	return str
};

utils.vizRule = function(svgSelection, rule, width, height) {

	function parseTargetChildren(target) {
		var children = [];

		if (!target) {
			return children;
		}

		if (target['$dive']) {
			for (var selector in target['$dive']) {
				var targetName = target['$dive'][selector]['$name'];
				var targetTemp = rule[targetName];

				children.push({
					name: targetName,
					desc: selector,
					type: 'target',
					children: parseTargetChildren(targetTemp),
				});
			}
		}

		if (target['$outputs'] && target['$outputs'].forEach) {
			target['$outputs'].forEach(function(output) {
				children.push({
					name: (output['$id'] ? '... ' : '') + output['$name'],
					type: output['$id'] ? 'listoutput' : 'objectoutput',
				});
			});
		}

		return children;
	}

	var treeData = {
		name: "Every " + (rule['$every'] || '1h'),
		desc: (rule['$startDay'] || rule['$startDayTime']) ? 'Next: ' + rule['$startDay'] + ' ' + rule['$startDayTime'] : 'No delay',
		children: [],
	};

	if (rule['$entries'] && rule['$entries'].forEach) {
		rule['$entries'].forEach(function(entry) {
			var targetName = entry['$name'];
			var targetTemp = rule[targetName];

			var node = {
				name: targetName,
				children: parseTargetChildren(targetTemp),
			};

			treeData.children.push(node);
		});
	}

	var svgGroup = d3.select(svgSelection).select('svg>g');
	if (svgGroup.empty()) {
		svgGroup = d3.select(svgSelection).append("svg")
			.attr("width", width)
			.attr("height", height)
			.call(d3.zoom().scaleExtent([0.5, 2]).on("zoom", function() {
				var transform = d3.event.transform;
				var translate = transform.x + ',' + transform.y;
				var scale = transform.k;
				svgGroup.attr("transform", "translate(" + translate + ")scale(" + scale + ")");
			}))
			.append("g");
	} else {
		svgGroup.selectAll('*').remove();

		d3.select(svgSelection).select('svg')
			.attr("width", width)
			.attr("height", height);
	}

	var tree = d3.tree().size([height, width]);

	var root = d3.hierarchy(treeData, function(d) { return d.children; });
	root.x0 = height / 2;
	root.y0 = 0;


	var idxSeq = 0;


	function diagonal(s, d) {
		path = `M ${s.y} ${s.x}
            C ${(s.y + d.y) / 2} ${s.x},
              ${(s.y + d.y) / 2} ${d.x},
              ${d.y} ${d.x}`

		return path
	}

	function click(d) {
		if (d.children) {
			d._children = d.children;
			d.children = null;
		} else {
			d.children = d._children;
			d._children = null;
		}
		update(d, 300);
	}

	function update(source, duration) {
		var treeData = tree(root);
		var nodes = treeData.descendants();
		var links = nodes.slice(1);

		nodes.forEach(function(d) { d.y = (d.depth + 1) * 180 });

		var node = svgGroup.selectAll('g.node')
			.data(nodes, function(d) { return d.id || (d.id = ++idxSeq); });

		var nodeEnter = node.enter().append('g')
			.attr('class', 'node')
			.attr("transform", function(d) {
				return "translate(" + source.y0 + "," + source.x0 + ")";
			})
			.on('click', click);

		nodeEnter.append('circle')
			.attr('cursor', 'pointer')
			.attr('class', 'node')
			.attr('r', 1e-6)
			.style('stroke', 'steelblue')
			.style('stroke-width', '2px')
			.style("fill", function(d) {
				return d._children ? "lightsteelblue" : "#fff";
			});

		nodeEnter.append('text') // Primary text
			.attr('cursor', 'pointer')
			.attr("dy", function(d) {
				return "0.3em";
			})
			.attr("x", function(d) {
				return d.children || d._children ? -13 : 13;
			})
			.attr("text-anchor", function(d) {
				return d.children || d._children ? "end" : "start";
			})
			.style("font-size", "12px")
			.style("fill", "#333")
			.text(function(d) {
				return d.data.name;
			});

		nodeEnter.append('text') // Secondary text
			.attr('cursor', 'pointer')
			.attr("dy", "1.8em")
			.attr("x", function(d) {
				return d.children || d._children ? -13 : 13;
			})
			.attr("text-anchor", function(d) {
				return d.children || d._children ? "end" : "start";
			})
			.style("font-size", "10px")
			.style("fill", "#999")
			.text(function(d) {
				return d.data.desc;
			});

		var nodeUpdate = nodeEnter.merge(node);

		nodeUpdate.transition()
			.duration(duration)
			.attr("transform", function(d) {
				return "translate(" + d.y + "," + d.x + ")";
			});

		nodeUpdate.select('circle.node')
			.attr('r', 6)
			.style("fill", function(d) {
				if (d.data.type == 'listoutput') {
					return "#eeee90";
				} else if (d.data.type == 'objectoutput') {
					return "#90ee90";
				}

				return d._children ? "lightsteelblue" : "#fff";
			});


		var nodeExit = node.exit().transition()
			.duration(duration)
			.attr("transform", function(d) {
				return "translate(" + source.y + "," + source.x + ")";
			})
			.remove();

		nodeExit.select('circle')
			.attr('r', 1e-6);

		nodeExit.select('text')
			.style('fill-opacity', 1e-6);


		var link = svgGroup.selectAll('path.link')
			.data(links, function(d) { return d.id; });

		var linkEnter = link.enter().insert('path', "g")
			.attr("class", "link")
			.attr('d', function(d) {
				var o = { x: source.x0, y: source.y0 }
				return diagonal(o, o)
			})
			.style('fill', 'none')
			.style('stroke', '#ddd')
			.style('stroke-width', '16px');

		var linkUpdate = linkEnter.merge(link);

		linkUpdate.transition()
			.duration(duration)
			.attr('d', function(d) { return diagonal(d, d.parent) });

		var linkExit = link.exit().transition()
			.duration(duration)
			.attr('d', function(d) {
				var o = { x: source.x, y: source.y }
				return diagonal(o, o)
			})
			.remove();

		nodes.forEach(function(d) {
			d.x0 = d.x;
			d.y0 = d.y;
		});
	}

	update(root, 0);
}