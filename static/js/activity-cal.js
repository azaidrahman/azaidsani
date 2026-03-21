(function() {
  var postSet = window.acalData.postSet;
  var events = window.acalData.events;

  var commitMap = {};
  var currentDate = new Date();
  var viewYear = currentDate.getFullYear();
  var viewMonth = currentDate.getMonth();

  // bounds: current month and 2 months back
  var maxYear = viewYear;
  var maxMonth = viewMonth;
  var minDate = new Date(currentDate.getFullYear(), currentDate.getMonth() - 2, 1);
  var minYear = minDate.getFullYear();
  var minMonth = minDate.getMonth();

  var gridEl = document.getElementById('acal-grid');
  var labelEl = document.getElementById('acal-label');
  var prevBtn = document.getElementById('acal-prev');
  var nextBtn = document.getElementById('acal-next');

  var monthNames = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
  var dayNames = ['Mo','Tu','We','Th','Fr','Sa','Su'];

  function pad(n) { return n < 10 ? '0' + n : '' + n; }

  function render() {
    while (gridEl.firstChild) gridEl.removeChild(gridEl.firstChild);

    labelEl.textContent = monthNames[viewMonth] + ' ' + viewYear;

    var counts = Object.values(commitMap);
    var maxCommits = counts.length ? Math.max.apply(null, counts) : 1;

    dayNames.forEach(function(dn) {
      var d = document.createElement('div');
      d.className = 'acal-dow';
      d.textContent = dn;
      gridEl.appendChild(d);
    });

    var daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate();
    var startDay = new Date(viewYear, viewMonth, 1).getDay() - 1;
    if (startDay < 0) startDay = 6;

    for (var e = 0; e < startDay; e++) {
      var empty = document.createElement('div');
      empty.className = 'acal-cell';
      gridEl.appendChild(empty);
    }

    for (var day = 1; day <= daysInMonth; day++) {
      var dateStr = viewYear + '-' + pad(viewMonth + 1) + '-' + pad(day);
      var commits = commitMap[dateStr] || 0;
      var hasPost = postSet[dateStr] || false;

      var cell = document.createElement('div');
      cell.className = 'acal-cell';

      if (commits > 0) {
        var minPct = 55;
        var maxPct = 90;
        var pct = minPct + (maxPct - minPct) * (commits / maxCommits);
        var circle = document.createElement('div');
        circle.className = 'git-circle';
        circle.style.width = pct + '%';
        circle.style.height = pct + '%';
        cell.appendChild(circle);
      }

      var num = document.createElement('span');
      num.className = 'acal-num';
      num.textContent = day;
      cell.appendChild(num);

      if (hasPost) {
        var dot = document.createElement('div');
        dot.className = 'post-dot';
        cell.appendChild(dot);
      }

      gridEl.appendChild(cell);
    }
  }

  function updateArrows() {
    prevBtn.disabled = (viewYear === minYear && viewMonth === minMonth);
    nextBtn.disabled = (viewYear === maxYear && viewMonth === maxMonth);
    prevBtn.style.opacity = prevBtn.disabled ? '0.25' : '1';
    nextBtn.style.opacity = nextBtn.disabled ? '0.25' : '1';
  }

  prevBtn.addEventListener('click', function() {
    if (viewYear === minYear && viewMonth === minMonth) return;
    viewMonth--;
    if (viewMonth < 0) { viewMonth = 11; viewYear--; }
    render();
    updateArrows();
  });

  nextBtn.addEventListener('click', function() {
    if (viewYear === maxYear && viewMonth === maxMonth) return;
    viewMonth++;
    if (viewMonth > 11) { viewMonth = 0; viewYear++; }
    render();
    updateArrows();
  });

  // Process GitHub events into commitMap
  if (Array.isArray(events)) {
    events.forEach(function(ev) {
      if (ev.type === 'PushEvent') {
        var date = ev.created_at.slice(0, 10);
        var count = (ev.payload && ev.payload.commits) ? ev.payload.commits.length : 1;
        commitMap[date] = (commitMap[date] || 0) + count;
      }
    });
  }
  render();
  updateArrows();
})();
