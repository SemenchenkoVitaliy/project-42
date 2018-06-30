const onload = function() {
  document.getElementsByTagName('iframe')[0].onload = () => {
    window.location.reload();
  };
};
