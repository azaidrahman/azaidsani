(function () {
  "use strict";

  var input = document.getElementById("tag-search");
  var results = document.getElementById("tag-search-results");
  var tags = window.sidebarTags || [];

  if (!input || !results) return;

  input.addEventListener("input", function () {
    var q = this.value.toLowerCase().trim();
    results.textContent = "";

    if (!q) return;

    var matches = tags.filter(function (t) {
      return t.name.toLowerCase().includes(q);
    });

    matches.slice(0, 5).forEach(function (t) {
      var a = document.createElement("a");
      a.href = t.url;
      a.className = "tag-pill small";
      a.textContent = t.name + " ";

      var span = document.createElement("span");
      span.className = "tag-count";
      span.textContent = t.count;
      a.appendChild(span);

      results.appendChild(a);
    });
  });
})();
