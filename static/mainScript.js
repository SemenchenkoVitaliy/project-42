const showNextTitle = () => {
  let cur = 0;
  return function(images){
    images[cur++].style.display = 'none';
    if(cur === images.length - 1) cur = 0;
    images[cur].style.display = 'block';
  }
}

const start = () => {
  const products = Array.from(document.getElementsByClassName('Product'));
  products.forEach(function(element, index) {
    let images = element.children[0].children;
    images[0].style.display = 'block';

    setTimeout(function(){
      let fn = showNextTitle();
      setInterval(function(){
        fn(images)
      }, 10000 * products.length)
    }, index * 2000)
  })
};
