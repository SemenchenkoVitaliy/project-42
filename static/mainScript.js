const showNextCover = () => {
  let cur = 0;
  return function(images) {
    images[cur++].style.display = 'none';
    if (cur === images.length) cur = 0;
    images[cur].style.display = 'block';
  };
};

const start = () => {
  const products = Array.from(document.getElementsByClassName('Product'));
  products.forEach((element, index) => {
    const images = Array.from(element.children[0].children);
    if (images[images.length - 1].tagName !== 'IMG') {
      images.pop();
    }
    images[0].style.display = 'block';

    setTimeout(() => {
      const fn = showNextCover();
      setInterval(() => {
        fn(images);
      }, 10000 * products.length);
    }, index * 2000);
  });
};
