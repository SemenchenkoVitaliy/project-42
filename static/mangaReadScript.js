
window.onkeydown = function(e) {
  return !(e.keyCode == 32);
};

document.onkeyup = function(e = window.event) {
  /*32 - space 37 - left 38 - up 39 - right 40 - down*/
  switch (e.keyCode) {
    case 32:
      changeImageSize();
      break;
    case 37:
      prevPage();
      break;
    case 39:
      nextPage();
      break;
    default:
      break;
  }
};

const getPageNum = () => {
  let hash = document.location.hash;
  return hash.replace('#', '') | 0;
}

const getCookie = () => {
  let result = {};

  document.cookie.split(';').forEach((cookie) => {
    cookie = cookie.trim();
    result[cookie.split('=')[0]] = parseInt(cookie.split('=')[1]);
  })
  return result
}

let ImageDisplay, images, imageSize = 0;

const start = () => {
  ImageDisplay = Array.from(document.getElementsByClassName('ImageDisplay'))[0];
  images = ImageDisplay.children[0].children;
  images[getPageNum()].style.display = 'block'

  const chapterIndex = parseInt(document.location.pathname.split('/')[3]);
  const pagesIndex = getPageNum();
  document.getElementById('chapters')[chapterIndex].selected = 'selected';
  document.getElementById('pages')[pagesIndex].selected = 'selected';

  let size = document.cookie.split(';')[0].split('=')[1];
  changeImageSize(parseInt(size));

  document.cookie = 'lastVisited=' + document.location.pathname + '; path=../';
}

const nextChapter = () => {
  const chapters = document.getElementById('chapters');
  if (chapters.selectedIndex + 1 === chapters.length) {
    const menu = document.getElementsByClassName('Menu')[0];
    document.location = menu.children[0].children[0].href;
  }
  document.location = chapters[chapters.selectedIndex + 1].value;
}

const prevChapter = () => {
  const chapters = document.getElementById('chapters');
  if (chapters.selectedIndex === 0) {
    const menu = document.getElementsByClassName('Menu')[0];
    document.location = menu.children[0].children[0].href;
  }
  document.location = chapters[chapters.selectedIndex - 1].value;
}

const nextPage = () => {
  let pages = document.getElementById('pages');
  let pageNum = getPageNum();

  if(pageNum + 1 === pages.length) nextChapter();

  document.getElementsByClassName('ImageDisplay')[0].scrollIntoView();
  document.location.hash = '#' + (pageNum + 1);
  document.cookie = 'lastPage=' + (pageNum + 1) + '; path=../';

  images[pageNum].style.display = 'none';
  images[pageNum + 1].style.display = 'block';
  pages[pageNum + 1].selected = 'selected';
}

const prevPage = () => {
  let pages = document.getElementById('pages');
  let pageNum = getPageNum();

  if(pageNum === 0) prevChapter();

  document.getElementsByClassName('ImageDisplay')[0].scrollIntoView();
  document.location.hash = '#' + (pageNum - 1);
  document.cookie = 'lastPage=' + (pageNum - 1) + '; path=../';

  images[pageNum].style.display = 'none';
  images[pageNum - 1].style.display = 'block';
  pages[pageNum - 1].selected = 'selected';
}

const changePage = (pageNum) => {
  document.location.hash = '#' + (pageNum - 1);
  document.cookie = 'lastPage=' + (pageNum - 1) + '; path=../';

  Array.from(images).forEach((image) => {
    image.style.display = 'none';
  });

  images[pageNum].style.display = 'block';
  pages[pageNum].selected = 'selected';
}

function changeImageSize(imgStat = imageSize) {
  const mainImages = ImageDisplay.children[0].children;
  switch (imgStat) {
    case 0:
      imageSize = 1;
      document.getElementById("size").textContent = "Size 50%";
      Array.from(mainImages).forEach((image) => {
        image.style.maxWidth = '50%';
        image.style.width = 'auto';
      })
      break;
    case 1:
      imageSize = 2;
      document.getElementById("size").textContent = "Size 100%";
      Array.from(mainImages).forEach((image) => {
        image.style.maxWidth = '10000%';
        image.style.width = '100%';
      })
      break;
    case 2:
      imageSize = 0;
      document.getElementById("size").textContent = "Size";
      Array.from(mainImages).forEach((image) => {
        image.style.maxWidth = '10000%';
        image.style.width = 'auto';
      })
      break;
    default:
      break;
  }
  document.cookie = 'size=' + imgStat + '; path=../';
}
