(function () {
  var overlay = document.getElementById("post-zoom");
  var zoomImg = document.getElementById("post-zoom-img");
  if (!overlay || !zoomImg) return;

  var isDesktop = window.matchMedia("(hover: hover) and (pointer: fine)").matches;
  if (!isDesktop) return;

  function openZoom(src, alt) {
    zoomImg.src = src;
    zoomImg.alt = alt || "";
    overlay.classList.add("active");
    document.body.style.overflow = "hidden";
  }

  function closeZoom() {
    overlay.classList.remove("active");
    zoomImg.src = "";
    document.body.style.overflow = "";
  }

  var images = document.querySelectorAll(".content__body img");
  for (var i = 0; i < images.length; i++) {
    images[i].addEventListener("click", function () {
      openZoom(this.src, this.alt);
    });
  }

  overlay.addEventListener("click", function () {
    closeZoom();
  });

  if (window._postZoomKeydown) {
    document.removeEventListener("keydown", window._postZoomKeydown);
  }

  window._postZoomKeydown = function (e) {
    if (e.key === "Escape" && overlay.classList.contains("active")) {
      closeZoom();
    }
  };

  document.addEventListener("keydown", window._postZoomKeydown);
})();
