// parus/web/terms/js/graph-term.js — граф для терминологической базы

import { ParseqGraphBase } from '../graph-render.js';
export class TermGraph extends ParseqGraphBase {
    constructor(containerId, termPath, graphParam = '', listParam = '') {
        super(containerId, '');
        this.termPath   = termPath;
        this.graphParam = graphParam;
        this.listParam  = listParam;}
// 
    async loadAndRender() {
        this.container.innerHTML = '<div class="dag-loading">load graph…</div>';
        try {
            const res = await fetch(`/api/terms/graph?path=${encodeURIComponent(this.termPath)}&t=${Date.now()}`);
            if (!res.ok) throw new Error('graph api error');
            const data = await res.json();
            this.centerIndex = data.center ? data.center.index : null;
            this.container.innerHTML = '';
            const param     = this.graphParam.trim().toLowerCase();
            const onlyList  = param === '+list';
            const onlySvg   = param === '-list';
            const skipSides = new Set(param.split(/\s+/).filter(s => ['-west','-nord','-east','-south'].includes(s)).map(s => s.slice(1)));
            if (skipSides.has('west'))  data.left   = [];
            if (skipSides.has('nord'))  data.top    = [];
            if (skipSides.has('east'))  data.right  = [];
            if (skipSides.has('south')) data.bottom = [];
            if (!onlyList) this._renderSvg(data);
            const hr = document.createElement('hr');
            hr.className = 'dag-divider';
            this.container.appendChild(hr);
            if (!onlySvg) this._renderList(data);
        } catch(e) {
            console.error(e);
            this.container.innerHTML = `<div class="dag-error">error graph: ${e.message}</div>`;}}
// 
    _renderList(data) {
        const param = (this.listParam || '').trim().toLowerCase();
        const skip = new Set(
            param.split(/\s+/)
            .filter(s => s.startsWith('-') && ['tags','links','includes','product',
            'inter-tags','inter-links','inter-includes','inter-product',
            'attribute','inter-attribute','preposition','inter-preposition',
            'prefix','suffix','inter-prefix','inter-suffix'].includes(s.slice(1)))
            .map(s => s.slice(1)));
        const sections = [
        { label: 'TG', nodes: data.tags,               cls: 'dag-section-nord',  key: 'tags'               },
        { label: 'LK', nodes: data.links,              cls: 'dag-section-west',  key: 'links'              },
        { label: 'IN', nodes: data.includes,           cls: 'dag-section-south', key: 'includes'           },
        { label: 'PD', nodes: data.product,            cls: 'dag-section-east',  key: 'product'            },
        { label: 'AT', nodes: data.attribute,          cls: 'dag-section-west',  key: 'attribute'          },
        { label: 'PP', nodes: data.preposition,        cls: 'dag-section-nord',  key: 'preposition'        },
        { label: 'IT', nodes: data.interTags,          cls: 'dag-section-nord',  key: 'inter-tags'         },
        { label: 'IL', nodes: data.interLinks,         cls: 'dag-section-west',  key: 'inter-links'        },
        { label: 'II', nodes: data.interIncludes,      cls: 'dag-section-south', key: 'inter-includes'     },
        { label: 'ID', nodes: data.interProduct,       cls: 'dag-section-east',  key: 'inter-product'      },
        { label: 'IA', nodes: data.interAttribute,     cls: 'dag-section-west',  key: 'inter-attribute'    },
        { label: 'IP', nodes: data.interPreposition,   cls: 'dag-section-nord',  key: 'inter-preposition'  },
        { label: 'PX', nodes: data.prefix,             cls: 'dag-section-west',  key: 'prefix'             },
        { label: 'SX', nodes: data.suffix,             cls: 'dag-section-east',  key: 'suffix'             },
        { label: 'IX', nodes: data.interPrefix,        cls: 'dag-section-west',  key: 'inter-prefix'       },
        { label: 'IS', nodes: data.interSuffix,        cls: 'dag-section-east',  key: 'inter-suffix'       },
        ].filter(s => !skip.has(s.key));
        const root = document.createElement('div');
        root.className = 'dag-list';
        sections.forEach(sec => {
            if (!sec.nodes || sec.nodes.length === 0) return;
            const div = document.createElement('div');
            div.className = 'dag-section ' + sec.cls;
            const label = document.createElement('span');
            label.className = 'dag-section-label';
            label.textContent = sec.label + ': ';
            div.appendChild(label);
            sec.nodes.forEach((n, i) => {
                if (i > 0) div.appendChild(document.createTextNode(' '));
                const a = this._makeLink(n);
                div.appendChild(a);
                const clean = (n.neighbors || []).filter(nb => nb.index !== this.centerIndex);
                if (clean.length > 0) {
                    div.appendChild(document.createTextNode(' [ '));
                    clean.forEach((nb, k) => {
                        if (k > 0) {
                            const sep = document.createElement('span');
                            sep.className = 'dag-link-sep';
                            sep.textContent = ' · ';
                            div.appendChild(sep);}
                        div.appendChild(this._makeLink(nb, true));});
                    div.appendChild(document.createTextNode(' ]'));}});
            root.appendChild(div);});
        this.container.appendChild(root);}
// 
    _makeLink(n, secondary = false) {
        const a = document.createElement('a');
        a.href = '#';
        a.textContent = n.title || n.index;
        if (n.hint) a.title = n.hint;
        a.className = secondary ? 'dag-link dag-link-secondary' : 'dag-link';
        a.onclick = e => {
            e.preventDefault();
            window.location.href = 'term.html?path=' + encodeURIComponent(n.index);};
        return a;}
// 
    _loadIcon(div, icon) {
        fetch('/api/terms/satellite-svg?path=' + encodeURIComponent(icon))
            .then(r => r.json())
            .then(d => {
                if (d.svg) {
                    div.innerHTML = d.svg;
                    const s = div.querySelector('svg');
                    if (s) s.setAttribute('style', 'width:100%;height:100%;');}})
            .catch(() => {});}
// 
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
        window.location.href = 'term.html?path=' + encodeURIComponent(nodeData.index);
        tip.remove();};
    tip.appendChild(header);
    if (clean.length > 0) {
        const ul = document.createElement('ul');
        ul.className = 'dag-tooltip-list';
        clean.forEach(n => {
            const li = document.createElement('li');
            const a = document.createElement('a');
            a.href = '#';
            a.textContent = n.title || n.index;
            a.onclick = e => {
                e.preventDefault();
                window.location.href = 'term.html?path=' + encodeURIComponent(n.index);
                tip.remove();};
            li.appendChild(a);
            ul.appendChild(li);});
        tip.appendChild(ul);}
    tip.style.left = (evt.clientX + 12) + 'px';
    tip.style.top  = evt.clientY + 'px';
    document.body.appendChild(tip);
    const close = () => {tip.remove(); document.removeEventListener('click', close);};
    setTimeout(() => document.addEventListener('click', close), 0);} }