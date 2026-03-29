(function() {
    var allKnownTags = [];

    // Fetch all existing tags once
    fetch('/api/tags')
        .then(function(r) { return r.json(); })
        .then(function(tags) { allKnownTags = tags || []; })
        .catch(function() { allKnownTags = []; });

    // Initialize all tag inputs on the page
    document.querySelectorAll('[data-tag-input]').forEach(initTagInput);

    function initTagInput(wrap) {
        var pillsContainer = wrap.querySelector('.tag-pills');
        var textInput = wrap.querySelector('.tag-text');
        var suggestionsDiv = wrap.querySelector('.tag-suggestions');
        var tags = [];

        // Load initial tags from data attribute
        var initial = wrap.dataset.initial;
        if (initial && initial.trim()) {
            initial.trim().split(/\s+/).forEach(function(t) {
                if (t) addTag(t);
            });
        }

        // Click on the wrap focuses the input
        wrap.addEventListener('click', function() { textInput.focus(); });

        textInput.addEventListener('input', function() {
            var val = textInput.value;

            // If user typed a space, commit the tag before the space
            if (val.indexOf(' ') !== -1) {
                var parts = val.split(/\s+/);
                parts.forEach(function(p) {
                    p = p.trim().toLowerCase();
                    if (p) addTag(p);
                });
                textInput.value = '';
                hideSuggestions();
                return;
            }

            showSuggestions(val.trim().toLowerCase());
        });

        textInput.addEventListener('keydown', function(e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                var val = textInput.value.trim().toLowerCase();
                if (val) addTag(val);
                textInput.value = '';
                hideSuggestions();
            }
            if (e.key === 'Backspace' && textInput.value === '' && tags.length > 0) {
                removeTag(tags[tags.length - 1]);
            }
        });

        function addTag(name) {
            name = name.trim().toLowerCase();
            if (!name || tags.indexOf(name) !== -1) return;
            tags.push(name);

            var pill = document.createElement('span');
            pill.className = 'pill editable';
            pill.dataset.tag = name;

            pill.appendChild(document.createTextNode(name));

            var hidden = document.createElement('input');
            hidden.type = 'hidden';
            hidden.name = 'tags';
            hidden.value = name;
            pill.appendChild(hidden);

            var x = document.createElement('button');
            x.type = 'button';
            x.className = 'remove-tag';
            x.textContent = 'x';
            x.addEventListener('click', function(e) {
                e.stopPropagation();
                removeTag(name);
            });
            pill.appendChild(x);

            pillsContainer.appendChild(pill);
        }

        function removeTag(name) {
            tags = tags.filter(function(t) { return t !== name; });
            var pill = pillsContainer.querySelector('[data-tag="' + name + '"]');
            if (pill) pill.remove();
        }

        function showSuggestions(query) {
            if (!query) { hideSuggestions(); return; }

            var matches = allKnownTags.filter(function(t) {
                return t.indexOf(query) === 0 && tags.indexOf(t) === -1;
            });

            if (matches.length === 0) { hideSuggestions(); return; }

            suggestionsDiv.textContent = '';
            matches.slice(0, 6).forEach(function(tag) {
                var btn = document.createElement('button');
                btn.type = 'button';
                btn.className = 'tag-suggestion-item';
                btn.textContent = tag;
                btn.addEventListener('click', function(e) {
                    e.stopPropagation();
                    addTag(tag);
                    textInput.value = '';
                    hideSuggestions();
                    textInput.focus();
                });
                suggestionsDiv.appendChild(btn);
            });
        }

        function hideSuggestions() {
            suggestionsDiv.textContent = '';
        }

        // Close suggestions on outside click
        document.addEventListener('click', function(e) {
            if (!wrap.contains(e.target)) hideSuggestions();
        });
    }

    // Also handle suggestion pills from the server (htmx tag-suggest partial)
    document.addEventListener('click', function(e) {
        var suggestion = e.target.closest('.suggestion');
        if (!suggestion) return;

        var tagName = suggestion.dataset.tag;
        if (!tagName) return;

        // Find the nearest tag input on the page and add to it
        var wrap = document.querySelector('[data-tag-input]');
        if (!wrap) return;

        var input = wrap.querySelector('.tag-text');
        // Trigger addTag by setting value and dispatching
        input.value = tagName + ' ';
        input.dispatchEvent(new Event('input'));
    });
})();
