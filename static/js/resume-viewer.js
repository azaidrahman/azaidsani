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
    if (e.key === "Escape" && overlay.style.display === "flex") {
      closePopup();
    }
  });
})();
