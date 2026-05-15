// parus/web/terms/js/lens-term-tags.js — линза тегов для терминологической базы

import { esc, buildCharOrder, compareByASort } from '../utils.js';
export function initTermTagsLens(ctx) {
  const { getPairs, getASort, onSave } = ctx;
  const frameIndex = ctx.frameIndex;
  let tagsCache    = [];
  let allTerms     = [];
// 
  document.getElementById('tagsToggleBtn').addEventListener('click', async () => {
    const container = document.getElementById('tagsContainer');
    const btn       = document.getElementById('tagsToggleBtn');
    if (!container.classList.contains('hidden')) {
      container.classList.add('hidden');
      btn.textContent = '⬡';
      btn.classList.remove('active');
      return;}
    container.classList.remove('hidden');
    btn.textContent = '⬡';
    btn.classList.add('active');
    await loadTagsPanel(container);});
// 
  async function loadTagsPanel(container) {
    try { tagsCache = await ctx.api.getTags(); }
    catch(e) { tagsCache = []; }
    try { allTerms = await ctx.api.getFrames(); }
    catch(e) { allTerms = []; }
    const aSort = getASort ? getASort() : '';
    if (aSort) {
      const order = buildCharOrder(aSort);
      tagsCache.sort((a, b) => compareByASort(a.title || '', b.title || '', order));}
    const tagsPair     = getPairs().find(p => p.Left === 'tags');
    const match = tagsPair ? tagsPair.Right.match(/\[([^\]]*)\]/) : null;
    const currentNames = match ? match[1].trim().split(/\s+/).filter(Boolean) : [];
    if (aSort) {
      const order = buildCharOrder(aSort);
      currentNames.sort((a, b) => compareByASort(a, b, order));}
    renderTagsPanel(container, currentNames);}
// 
  function renderTagsPanel(container, currentNames) {
    container.innerHTML = `
      <div class="tags-panel">
        <div class="tags-panel-pills" id="tagsPills"></div>
        <div class="tags-panel-input-wrap">
          <input class="tags-panel-input" id="tagsInput"
            type="text" placeholder="add tag..."
            autocomplete="off" spellcheck="false">
          <div class="tags-autocomplete" id="tagsAutocomplete"></div>
        </div>
      </div>`;
    const pillsEl  = document.getElementById('tagsPills');
    const input    = document.getElementById('tagsInput');
    const dropdown = document.getElementById('tagsAutocomplete');
// 
    function makePill(name) {
      const tag   = tagsCache.find(t => t.title.toLowerCase() === name.toLowerCase());
      const color = tag?.color || '';
      const icon  = tag?.icon  || '';
      const title = tag ? tag.title : name;
      const path  = tag ? tag.index : '';
      const pill  = document.createElement('span');
      pill.className    = 'tag-pill';
      pill.dataset.name = name;
      if (color) pill.style.borderColor = color;
      pill.innerHTML =
          `<span class="tag-pill-icon-svg"></span>`
        + `<span class="tag-pill-title">${esc(title)}</span>`
        + `<span class="tag-pill-expand" title="members">▾</span>`
        + `<span class="tag-pill-remove">×</span>`;
      pill.dataset.path = path;
// 
      pill.querySelector('.tag-pill-remove').addEventListener('click', e => {
        e.stopPropagation();
        pill.remove();
        autoSave();});
      pill.querySelector('.tag-pill-title').addEventListener('click', () => {
        if (path) window.location.href = 'term.html?path=' + encodeURIComponent(path);});
      pill.querySelector('.tag-pill-expand').addEventListener('click', async e => {
        e.stopPropagation();
        const existing = pill.querySelector('.tag-pill-list');
        if (existing) {existing.remove(); return;}
        document.querySelectorAll('.tag-pill-list').forEach(el => el.remove());
        const list = document.createElement('div');
        list.className = 'tag-pill-list';
        if (!path) {
          list.innerHTML = '<span class="tag-pill-list-empty">not list</span>';
        } else {
          const result = await ctx.api.getChildren(path, '', getASort ? getASort() : '');
          if (!result || result.length === 0) {
            list.innerHTML = '<span class="tag-pill-list-empty">not list</span>';
          } else {
            result.forEach(c => {
              const item = document.createElement('a');
              item.className = 'tag-pill-list-item';
              item.href = 'term.html?path=' + encodeURIComponent(c.index);
              item.innerHTML =
                  `<span class="tag-pill-list-marker">▪</span>`
                + `<span class="tag-pill-list-title">${esc(c.name || c.title)}</span>`;
              list.appendChild(item);});}}
        pill.appendChild(list);});
      return pill;}
// 
    function renderPills(names) {
      pillsEl.innerHTML = '';
      names.forEach(n => {
        const pill = makePill(n);
        pillsEl.appendChild(pill);
        const path = pill.dataset.path;
        if (path) {
          fetch('/api/terms/satellite-svg?path=' + encodeURIComponent(path))
            .then(r => r.json())
            .then(d => {
              if (d.svg) {
                const el = pill.querySelector('.tag-pill-icon-svg');
                if (el) el.innerHTML = d.svg;}})
            .catch(() => {});}});}
// 
    function addPill(name) {
      const used = new Set([...pillsEl.querySelectorAll('.tag-pill')].map(p => p.dataset.name));
      if (!used.has(name)) pillsEl.appendChild(makePill(name));
      autoSave();}
// 
    renderPills(currentNames);
// 
    async function autoSave() {
      const names = [...pillsEl.querySelectorAll('.tag-pill')].map(p => p.dataset.name);
      const fresh = await fetch('/api/terms/' + encodeURIComponent(frameIndex));
      const freshData = await fresh.json();
      const freshPairs = freshData.pairs || [];
      const pairs = getPairs();
      for (let p of pairs) {
        if (p.Left === 'tags') {
            const freshPair = freshPairs.find(fp => fp.Left === 'tags');
            const current = freshPair ? freshPair.Right : p.Right;
            const match = current.match(/\[([^\]]*)\]\s*\[([^\]]*)\]/);
            const inverse = match ? match[2].trim() : '';
            p.Right = '[ ' + names.join(' ') + ' ] [ ' + inverse + ' ]';}
//
        if (p.Left === 'modified') {p.Right = new Date().toISOString().slice(0, 10);}}
      await fetch('/api/terms/' + encodeURIComponent(frameIndex), {
        method: 'PUT',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({pairs})});
      const freshResult = await fetch('/api/terms/' + encodeURIComponent(frameIndex));
      const data  = await freshResult.json();
      onSave(data.pairs || []);}
// 
    input.addEventListener('input', () => {
      const val = input.value.toLowerCase().trim();
      dropdown.innerHTML = '';
      if (!val) {dropdown.style.display = 'none'; return;}
      const used = new Set([...pillsEl.querySelectorAll('.tag-pill')].map(p => p.dataset.name));
      const matches = tagsCache.filter(t =>
        t.title.toLowerCase().includes(val) && !used.has(t.title));
      dropdown.style.display = 'block';
      matches.slice(0, 26).forEach(tag => {
        const item = document.createElement('div');
        item.className = 'tags-autocomplete-item';
        item.innerHTML = `<span>${esc(tag.title)}</span>`;
        item.onmousedown = e => {
          e.preventDefault();
          addPill(tag.title);
          input.value = '';
          dropdown.style.display = 'none';};
        dropdown.appendChild(item);});
      const existingTerm = allTerms.find(t =>
        t.title.toLowerCase() === val &&
        !tagsCache.find(tc => tc.title.toLowerCase() === t.title.toLowerCase()));
      if (existingTerm) {
        const makeTagItem = document.createElement('div');
        makeTagItem.className = 'tags-autocomplete-item tags-autocomplete-new';
        makeTagItem.innerHTML = `<span>сделать тегом «${esc(existingTerm.title)}»</span>`;
        makeTagItem.onmousedown = async e => {
          e.preventDefault();
          input.value    = '';
          input.disabled = true;
          dropdown.style.display = 'none';
          try {
            const freshTerm = await fetch('/api/terms/' + encodeURIComponent(existingTerm.index));
            const termData  = await freshTerm.json();
            const newPairs  = termData.pairs.map(p => p.Left === 'rank' ? {...p, Right: '0'} : p);
            await fetch('/api/terms/' + encodeURIComponent(existingTerm.index), {
              method: 'PUT',
              headers: {'Content-Type': 'application/json'},
              body: JSON.stringify({pairs: newPairs})});
            tagsCache.push({index: existingTerm.index, title: existingTerm.title, color: '', icon: ''});
            addPill(existingTerm.title);
          } catch(err) {console.error('make tag:', err);}
          input.disabled = false;};
        dropdown.appendChild(makeTagItem);}
      if (matches.length === 0 && !existingTerm) {
        const newItem = document.createElement('div');
        newItem.className = 'tags-autocomplete-item tags-autocomplete-new';
        newItem.innerHTML = `<span>создать тег «${esc(input.value.trim())}»</span>`;
        newItem.onmousedown = async e => {
          e.preventDefault();
          const newName = input.value.trim();
          if (!newName) return;
          input.value    = '';
          input.disabled = true;
          dropdown.style.display = 'none';
          try {
            const r = await fetch('/api/terms', {
              method: 'POST',
              headers: {'Content-Type': 'application/json'},
              body: JSON.stringify({name: newName})});
            const d = await r.json();
            if (d.path) {
              const freshTerm = await fetch('/api/terms/' + encodeURIComponent(d.path));
              const termData  = await freshTerm.json();
              const newPairs  = termData.pairs.map(p => p.Left === 'rank' ? {...p, Right: '0'} : p);
              await fetch('/api/terms/' + encodeURIComponent(d.path), {
                method: 'PUT',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({pairs: newPairs})});
              tagsCache.push({index: d.path, title: newName, color: '', icon: ''});
              addPill(newName);}
          } catch(err) {console.error('create tag:', err);}
          input.disabled = false;};
        dropdown.appendChild(newItem);}});
// 
    input.addEventListener('blur', () => {
      setTimeout(() => {dropdown.style.display = 'none';}, 150);});} }