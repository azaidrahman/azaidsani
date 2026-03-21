(function () {
  var overlay = document.getElementById("resume-popup");
  var openBtn = document.getElementById("download-btn");
  var closeBtn = document.getElementById("popup-close");

  function openPopup() {
    overlay.style.display = "flex";
  }

  function closePopup() {
    overlay.style.display = "none";
  }

  openBtn.addEventListener("click", openPopup);
  closeBtn.addEventListener("click", closePopup);

  overlay.addEventListener("click", function (e) {
    if (e.target === overlay) {
      closePopup();
    }
  });

  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape") {
      if (zoomOverlay.classList.contains("active")) {
        closeZoom();
      } else if (overlay.style.display === "flex") {
        closePopup();
      }
    }
  });

  // --- Click-to-zoom (desktop only) ---
  var zoomOverlay = document.getElementById("resume-zoom");
  var zoomImg = document.getElementById("resume-zoom-img");
  var isDesktop = window.matchMedia("(hover: hover) and (pointer: fine)").matches;

  function openZoom(src, alt) {
    zoomImg.src = src;
    zoomImg.alt = alt;
    zoomOverlay.classList.add("active");
    document.body.style.overflow = "hidden";
    // Center scroll after image loads
    zoomImg.onload = function () {
      zoomOverlay.scrollLeft = (zoomOverlay.scrollWidth - zoomOverlay.clientWidth) / 2;
      zoomOverlay.scrollTop = 0;
    };
  }

  function closeZoom() {
    zoomOverlay.classList.remove("active");
    zoomImg.src = "";
    document.body.style.overflow = "";
  }

  if (isDesktop) {
    var pageImages = document.querySelectorAll(".resume-page-img");
    for (var i = 0; i < pageImages.length; i++) {
      pageImages[i].addEventListener("click", function () {
        openZoom(this.src, this.alt);
      });
    }
  }

  // Close on click, but not if user was scrolling/dragging
  var didDrag = false;
  zoomOverlay.addEventListener("mousedown", function () { didDrag = false; });
  zoomOverlay.addEventListener("mousemove", function () { didDrag = true; });
  zoomOverlay.addEventListener("mouseup", function (e) {
    if (!didDrag) closeZoom();
  });
})();
