// parus/web/js/graph.js — граф для обычных баз и мета-базы

import { ParseqGraphBase } from './graph-render.js';

export class ParseqGraph extends ParseqGraphBase {
    constructor(containerId, projectId, graphParam = '', listParam = '', westSpec = '', nordSpec = '') {
        super(containerId, projectId);
        this.graphParam = graphParam;
        this.listParam  = listParam;
        this.westSpec   = westSpec;
        this.nordSpec   = nordSpec;}

    async loadAndRender(frameIndex) {
        this.container.innerHTML = '<div class="dag-loading">load graph…</div>';
        try {
            const res = await fetch(`/api/projects/${this.projectId}/graph/${frameIndex}?t=${Date.now()}`);
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

    _renderList(data) {
        const param = (this.listParam || '').trim().toLowerCase();
        const skip = new Set(
            param.split(/\s+/)
            .filter(s => s.startsWith('-') && ['links','tags','includes','product','top','bottom'].includes(s.slice(1)))
            .map(s => s.slice(1)));
        const isMeta = this.projectId === '00000-META-BASE';
        const sections = [
            { label: 'L', nodes: data.links,    cls: 'dag-section-west',  key: 'links'   },
            { label: 'T', nodes: data.tags,      cls: 'dag-section-nord',  key: 'tags'    },
            { label: 'I', nodes: data.includes,  cls: 'dag-section-south', key: 'includes'},
            { label: 'P', nodes: data.product,   cls: 'dag-section-east',  key: 'product' },
            ...(isMeta ? [
                { label: 'N', nodes: data.top,    cls: 'dag-section-nord',  key: 'top'   },
                { label: 'S', nodes: data.bottom, cls: 'dag-section-south', key: 'bottom'},
            ] : []),
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
        this.container.appendChild(root);}}

window.ParseqGraph = ParseqGraph;