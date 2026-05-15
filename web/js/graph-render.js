// parus/web/js/graph-render.js — базовый класс рендеринга графа

export class ParseqGraphBase {
    constructor(containerId, projectId) {
        this.container = document.getElementById(containerId);
        this.projectId = projectId;
        this.svgNS = 'http://www.w3.org/2000/svg';
        this.centerIndex = null;
        this.g = {
            padding:      24,
            nodeW:        80,
            nodeH:        40,
            vSpacing:     10,
            hSpacing:     32,
            collectorGap: 24,};}

    _loadIcon(div, icon) {
        fetch('/api/projects/' + this.projectId + '/content/' + icon)
            .then(r => r.json())
            .then(d => {
                if (d.contentType === 'svg' && d.contentRaw) {
                    div.innerHTML = d.contentRaw;
                    const s = div.querySelector('svg');
                    if (s) s.setAttribute('style', 'width:100%;height:100%;');}})
            .catch(() => {});}

    _renderSvg(data) {
        const g = this.g;
        const has = {
            left:    data.left    && data.left.length    > 0,
            top:     data.top     && data.top.length     > 0,
            right:   data.right   && data.right.length   > 0,
            bottom:  data.bottom  && data.bottom.length  > 0,};
        has.incoming = has.left || has.top;
        const colH = n => n > 0 ? n * g.nodeH + (n - 1) * g.vSpacing : 0;
        const h = {
            left:   colH(data.left.length),
            top:    colH(data.top.length),
            right:  colH(data.right.length),
            bottom: colH(data.bottom.length),};
        const c = {};
        let x = g.padding;
        if (has.left) { c.x_left = x; x += g.nodeW + g.hSpacing; c.x_left_col = x; }
        if (has.incoming) { x = (c.x_left_col || g.padding) + (has.top ? g.collectorGap : 0); c.x_merge = x; c.x_tb_col = x; }
        x += g.hSpacing; c.x_center = x; x += g.nodeW;
        if (!c.x_merge) { c.x_merge = c.x_center; }
        x += g.hSpacing; c.x_bottom_col = x;
        if (has.right) { x += g.collectorGap; c.x_right_col = x; x += g.hSpacing; c.x_right = x; x += g.nodeW; }
        const totalW = (has.bottom && !has.right ? c.x_bottom_col : x) + g.padding;
        const topH    = has.top    ? h.top    + g.vSpacing * 2 : 0;
        const botH    = has.bottom ? h.bottom + g.vSpacing * 2 : 0;
        const reqAbove = Math.max(h.left / 2 || 0, h.right / 2 || 0, topH + g.nodeH / 2);
        const reqBelow = Math.max(h.left / 2 || 0, h.right / 2 || 0, botH + g.nodeH / 2);
        c.y_axis       = reqAbove + g.padding;
        c.y_center_top = c.y_axis - g.nodeH / 2;
        c.y_top_start  = c.y_center_top - topH;
        c.y_bot_start  = c.y_center_top + g.nodeH + (has.bottom ? g.vSpacing * 2 : 0);
        c.y_left_start = c.y_axis - h.left  / 2;
        c.y_right_start= c.y_axis - h.right / 2;
        const totalH = reqAbove + reqBelow + g.padding * 2;
        const svg = document.createElementNS(this.svgNS, 'svg');
        svg.setAttribute('width',  totalW);
        svg.setAttribute('height', totalH);
        svg.setAttribute('class',  'dag-svg');
        const defs   = document.createElementNS(this.svgNS, 'defs');
        const marker = document.createElementNS(this.svgNS, 'marker');
        marker.setAttribute('id',           'dag-arrow');
        marker.setAttribute('viewBox',      '0 0 10 10');
        marker.setAttribute('refX',         '8');
        marker.setAttribute('refY',         '5');
        marker.setAttribute('markerWidth',  '5');
        marker.setAttribute('markerHeight', '5');
        marker.setAttribute('orient',       'auto-start-reverse');
        const arrowPath = document.createElementNS(this.svgNS, 'path');
        arrowPath.setAttribute('d',    'M 0 0 L 10 5 L 0 10 z');
        arrowPath.setAttribute('fill', 'var(--ink3, #a8a49e)');
        marker.appendChild(arrowPath);
        defs.appendChild(marker);
        svg.appendChild(defs);
        this.container.appendChild(svg);
        const poly = (pts, arrow, colorVar) => {
            const el = document.createElementNS(this.svgNS, 'polyline');
            el.setAttribute('points', pts.map(p => `${p.x},${p.y}`).join(' '));
            el.setAttribute('stroke', colorVar
                ? `var(${colorVar}, var(--border, #d4cfc7))`
                : 'var(--border, #d4cfc7)');
            el.setAttribute('stroke-width', '1.5');
            el.setAttribute('fill', 'none');
            if (arrow) el.setAttribute('marker-end', 'url(#dag-arrow)');
            svg.appendChild(el);};
        const node = (n, x, y, isCenter, colorVar) => {
            const grp = document.createElementNS(this.svgNS, 'g');
            grp.setAttribute('transform', `translate(${x},${y})`);
            grp.setAttribute('class', isCenter ? 'dag-node dag-node-center' : 'dag-node');
            const rect = document.createElementNS(this.svgNS, 'rect');
            rect.setAttribute('width',  g.nodeW);
            rect.setAttribute('height', g.nodeH);
            rect.setAttribute('rx', 3);
            rect.setAttribute('class', 'dag-node-rect');
            if (n.color) {
                rect.style.stroke = n.color;
                rect.style.fill   = 'transparent';
            } else if (colorVar) {
                rect.style.stroke = `var(${colorVar})`;
                rect.style.fill   = `var(${colorVar}-bg, transparent)`;}
            grp.appendChild(rect);
            if (n.icon) {
                const iconSize = g.nodeH - 6;
                const fo = document.createElementNS(this.svgNS, 'foreignObject');
                fo.setAttribute('x', 2);
                fo.setAttribute('y', 2);
                fo.setAttribute('width', iconSize);
                fo.setAttribute('height', iconSize);
                fo.style.pointerEvents = 'none';
                const div = document.createElement('div');
                div.style.cssText = 'width:100%;height:100%;display:flex;align-items:center;justify-content:flex-start;';
//                if (n.icon.includes('.')) {
                  if (n.icon.includes('.') && !n.icon.endsWith('.term')) {
                    div.innerHTML = `<img src="/img/${this.projectId}/${n.icon}" style="width:100%;height:100%;object-fit:contain;">`;
//                } else {
//                    fetch('/api/projects/' + this.projectId + '/content/' + n.icon)
//                        .then(r => r.json())
//                        .then(d => {
//                            if (d.contentType === 'svg' && d.contentRaw) {
//                                div.innerHTML = d.contentRaw;
//                                const s = div.querySelector('svg');
//                                if (s) s.setAttribute('style', 'width:100%;height:100%;');}})
//                        .catch(() => {});}
                } else {
                    this._loadIcon(div, n.icon);}
                fo.appendChild(div);
                grp.appendChild(fo);}
            const tx = g.nodeW - 10;
            const ty = g.nodeH / 2;
            const tgrp = document.createElementNS(this.svgNS, 'g');
            tgrp.setAttribute('class', 'dag-trigger');
            const hit = document.createElementNS(this.svgNS, 'rect');
            hit.setAttribute('x',      g.nodeW - 18);
            hit.setAttribute('y',      0);
            hit.setAttribute('width',  18);
            hit.setAttribute('height', g.nodeH);
            hit.setAttribute('fill',   'transparent');
            hit.setAttribute('cursor', 'pointer');
            const arr = document.createElementNS(this.svgNS, 'path');
            if (isCenter) {
                arr.setAttribute('d', `M ${tx-4} ${ty-4} L ${tx+2} ${ty} L ${tx-4} ${ty+4} z`);
                arr.setAttribute('class', 'dag-trigger-arrow active');
                tgrp.appendChild(hit);
                tgrp.appendChild(arr);
                tgrp.addEventListener('click', e => {
                    e.stopPropagation();
                    this._toggleTooltip(e, n);});
            } else {
                arr.setAttribute('d', `M ${tx-4} ${ty-4} L ${tx+2} ${ty} L ${tx-4} ${ty+4} z`);
                arr.setAttribute('class', n.neighbors && n.neighbors.length > 0
                    ? 'dag-trigger-arrow active'
                    : 'dag-trigger-arrow');
                tgrp.appendChild(hit);
                tgrp.appendChild(arr);
                tgrp.addEventListener('click', e => {
                    e.stopPropagation();
                    this._toggleTooltip(e, n);});
                grp.addEventListener('click', () => {
                    document.dispatchEvent(new CustomEvent('parseq-navigate', {
                        detail: { projectId: this.projectId, frameIndex: n.index }}));});}
            grp.appendChild(tgrp);
            svg.appendChild(grp);};
        if (has.left) {
            const y1 = c.y_left_start + g.nodeH / 2;
            const y2 = y1 + h.left - g.nodeH;
            if (y2 > y1) poly([{x: c.x_left_col, y: y1}, {x: c.x_left_col, y: y2}], false, '--graph-west');
            data.left.forEach((n, i) => {
                const y = c.y_left_start + i * (g.nodeH + g.vSpacing);
                node(n, c.x_left, y, false, '--graph-west');
                poly([{x: c.x_left + g.nodeW, y: y + g.nodeH / 2}, {x: c.x_left_col, y: y + g.nodeH / 2}], false, '--graph-west');});
            poly([{x: c.x_left_col, y: c.y_axis}, {x: c.x_merge, y: c.y_axis}], false, '--graph-west');}
        if (has.top) {
            const y1 = c.y_top_start + g.nodeH / 2;
            const y2 = y1 + h.top - g.nodeH;
            if (y2 > y1) poly([{x: c.x_tb_col, y: y1}, {x: c.x_tb_col, y: y2}], false, '--graph-nord');
            data.top.forEach((n, i) => {
                const y = c.y_top_start + i * (g.nodeH + g.vSpacing);
                node(n, c.x_center, y, false, '--graph-nord');
                poly([{x: c.x_center, y: y + g.nodeH / 2}, {x: c.x_tb_col, y: y + g.nodeH / 2}], false, '--graph-nord');});
            poly([{x: c.x_tb_col, y: y2}, {x: c.x_tb_col, y: c.y_axis}, {x: c.x_merge, y: c.y_axis}], false, '--graph-nord');}
        if (has.bottom) {
            const y1 = c.y_bot_start + g.nodeH / 2;
            const y2 = y1 + h.bottom - g.nodeH;
            if (y2 > y1) poly([{x: c.x_bottom_col, y: y1}, {x: c.x_bottom_col, y: y2}], false, '--graph-south');
            data.bottom.forEach((n, i) => {
                const y = c.y_bot_start + i * (g.nodeH + g.vSpacing);
                node(n, c.x_center, y, false, '--graph-south');
                poly([{x: c.x_bottom_col, y: y + g.nodeH / 2}, {x: c.x_center + g.nodeW, y: y + g.nodeH / 2}], true, '--graph-south');});
            poly([{x: c.x_bottom_col, y: y1}, {x: c.x_bottom_col, y: c.y_axis}, {x: c.x_center + g.nodeW, y: c.y_axis}], false, '--graph-south');}
        if (has.incoming) {
            poly([{x: c.x_merge, y: c.y_axis}, {x: c.x_center, y: c.y_axis}], true);}
        node(data.center, c.x_center, c.y_center_top, true);
        if (has.right) {
            poly([{x: c.x_center + g.nodeW, y: c.y_axis}, {x: c.x_right_col, y: c.y_axis}], false, '--graph-east');
            const y1 = c.y_right_start + g.nodeH / 2;
            const y2 = y1 + h.right - g.nodeH;
            if (y2 > y1) poly([{x: c.x_right_col, y: y1}, {x: c.x_right_col, y: y2}], false, '--graph-east');
            data.right.forEach((n, i) => {
                const y = c.y_right_start + i * (g.nodeH + g.vSpacing);
                node(n, c.x_right, y, false, '--graph-east');
                poly([{x: c.x_right_col, y: y + g.nodeH / 2}, {x: c.x_right, y: y + g.nodeH / 2}], true, '--graph-east');});}}

    _makeLink(n, secondary = false) {
        const a = document.createElement('a');
        a.href = '#';
        a.textContent = n.title || n.index;
        if (n.hint) a.title = n.hint;
        a.className = secondary ? 'dag-link dag-link-secondary' : 'dag-link';
        a.onclick = e => {
            e.preventDefault();
            if (n.index && n.index.startsWith('base:')) {
                window.location.href = 'project.html?id=' + n.index.slice(5);
                return;}
            document.dispatchEvent(new CustomEvent('parseq-navigate', {
                detail: { projectId: this.projectId, frameIndex: n.index }}));};
        return a;}

    _toggleTooltip(evt, nodeData) {
        document.querySelectorAll('.dag-tooltip').forEach(el => el.remove());
        const clean = (nodeData.neighbors || []).filter(n => n.index !== this.centerIndex);
        const tip = document.createElement('div');
        tip.className = 'dag-tooltip';
        const header = document.createElement('a');
        header.href = '#';
        header.textContent = nodeData.title || nodeData.index;
        header.className = 'dag-tooltip-title';
        header.onclick = e => {
            e.preventDefault();
            if (nodeData.index && nodeData.index.startsWith('base:')) {
                window.location.href = 'project.html?id=' + nodeData.index.slice(5);
                tip.remove();
                return;}
            document.dispatchEvent(new CustomEvent('parseq-navigate', {
                detail: { projectId: this.projectId, frameIndex: nodeData.index }}));
            tip.remove();};
        tip.appendChild(header);
        if (clean.length > 0) {
            const ul = document.createElement('ul');
            ul.className = 'dag-tooltip-list';
            clean.forEach(n => {
                const li = document.createElement('li');
                const a  = document.createElement('a');
                a.href = '#';
                a.textContent = n.title || n.index;
                a.onclick = e => {
                    e.preventDefault();
                    if (n.index && n.index.startsWith('base:')) {
                        window.location.href = 'project.html?id=' + n.index.slice(5);
                        tip.remove();
                        return;}
                    document.dispatchEvent(new CustomEvent('parseq-navigate', {
                        detail: { projectId: this.projectId, frameIndex: n.index }}));
                    tip.remove();};
                li.appendChild(a);
                ul.appendChild(li);});
            tip.appendChild(ul);}
        tip.style.left = (evt.clientX + 12) + 'px';
        tip.style.top  = evt.clientY + 'px';
        document.body.appendChild(tip);
        const close = () => { tip.remove(); document.removeEventListener('click', close); };
        setTimeout(() => document.addEventListener('click', close), 0);}}