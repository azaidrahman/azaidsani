document.addEventListener('DOMContentLoaded', function () {
  if (typeof Swup === 'undefined') {
    console.warn('swup: library not loaded, falling back to standard navigation');
    return;
  }

  var swup = new Swup({
    containers: ['#swup-body', '#swup-aside'],
    plugins: [
      new SwupHeadPlugin(),
      new SwupScriptsPlugin()
    ]
  });

  // Toggle body class for homepage vs non-homepage styling
  function updateBodyClass() {
    var isHome = window.location.pathname === '/' || window.location.pathname === '/index.html';
    if (isHome) {
      document.body.classList.remove('not-home');
    } else {
      document.body.classList.add('not-home');
    }
  }

  // Scroll to top and update body class on forward navigation
  swup.hooks.on('page:view', function () {
    updateBodyClass();
    window.scrollTo(0, 0);
  });
});
