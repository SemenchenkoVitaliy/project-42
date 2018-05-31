const startKeyHandler = function() {
  let scroller = null;
  const scrollSpeed = 5;
  const keys = {
    SPACEBAR: 32,
    LEFT_ARROW: 37,
    UP_ARROW: 38,
    RIGHT_ARROW: 39,
    DOWN_ARROW: 40,
    A: 65,
    D: 68,
    F: 70,
    S: 83,
    W: 87
  };

  document.onkeydown = function(e) {
    clearInterval(scroller);
    scroller = null;

    switch (e.which) {
      case keys.DOWN_ARROW:
      case keys.S:
        scroller = setInterval(() => {
          document.scrollingElement.scrollTop += scrollSpeed;
        }, 5);
        e.preventDefault();
        break;
      case keys.UP_ARROW:
      case keys.W:
        scroller = setInterval(() => {
          document.scrollingElement.scrollTop -= scrollSpeed;
        }, 5);
        e.preventDefault();
        break;
    }
  };

  document.onkeyup = function(e = window.event) {
    clearInterval(scroller);
    scroller = null;
    switch (e.keyCode) {
      case keys.F:
        changeLayout();
        break;
      case keys.SPACEBAR:
        changeImageSize();
        break;
      case keys.LEFT_ARROW:
      case keys.A:
        prevPage();
        break;
      case keys.RIGHT_ARROW:
      case keys.D:
        nextPage();
        break;
      default:
        break;
    }
  };

  window.onkeydown = function(e) {
    return !(e.keyCode === 32);
  };
};

const getPageNum = () => {
  const hash = document.location.hash;
  return hash.replace('#', '') | 0;
};

const getCookie = () => {
  const result = {};

  document.cookie.split(';').forEach((cookie) => {
    cookie = cookie.trim();
    result[cookie.split('=')[0]] = parseInt(cookie.split('=')[1]);
  });
  return result;
};

let images = {},
    pages = {},
    imageSize = 0,
    imageLayout = 0;

const nextPage = () => {
  const pageNum = getPageNum();

  if (pageNum + 1 === pages.length) nextChapter();

  document.location.hash = '#' + (pageNum + 1);
  document.cookie = 'lastPage=' + (pageNum + 1) + '; path=../';

  if (!imageLayout) images[pageNum].style.display = 'none';
  images[pageNum + 1].style.display = 'block';
  pages[pageNum + 1].selected = 'selected';
  images[pageNum + 1].scrollIntoView({ behavior: 'smooth' });
};

const prevPage = () => {
  const pageNum = getPageNum();

  if (pageNum === 0) prevChapter();

  document.location.hash = '#' + (pageNum - 1);
  document.cookie = 'lastPage=' + (pageNum - 1) + '; path=../';

  if (!imageLayout) images[pageNum].style.display = 'none';
  images[pageNum - 1].style.display = 'block';
  pages[pageNum - 1].selected = 'selected';
  images[pageNum - 1].scrollIntoView({ behavior: 'smooth' });
};

const nextChapter = () => {
  const chapters = document.getElementById('chapters');
  if (chapters.selectedIndex + 1 === chapters.length) {
    const menu = document.getElementsByClassName('Menu')[0];
    document.location = menu.children[0].children[0].href;
  }
  document.location = chapters[chapters.selectedIndex + 1].value;
};

const prevChapter = () => {
  const chapters = document.getElementById('chapters');
  if (chapters.selectedIndex === 0) {
    const menu = document.getElementsByClassName('Menu')[0];
    document.location = menu.children[0].children[0].href;
  }
  document.location = chapters[chapters.selectedIndex - 1].value;
};

const changePage = (pageNum) => {
  document.location.hash = '#' + (pageNum - 1);
  document.cookie = 'lastPage=' + (pageNum - 1) + '; path=../';

  Array.from(images).forEach((image) => {
    image.style.display = 'none';
  });

  images[pageNum].style.display = 'block';
  pages[pageNum].selected = 'selected';
};

function changeImageSize(imgStat = imageSize) {
  switch (imgStat) {
    case 0:
      imageSize = 1;
      document.getElementById('size').textContent = 'Size 50%';
      Array.from(images).forEach((image) => {
        image.style.maxWidth = '50%';
        image.style.width = 'auto';
      });
      break;
    case 1:
      imageSize = 2;
      document.getElementById('size').textContent = 'Size 100%';
      Array.from(images).forEach((image) => {
        image.style.maxWidth = '10000%';
        image.style.width = '100%';
      });
      break;
    case 2:
      imageSize = 0;
      document.getElementById('size').textContent = 'Size';
      Array.from(images).forEach((image) => {
        image.style.maxWidth = '10000%';
        image.style.width = 'auto';
      });
      break;
    default:
      break;
  }
  document.cookie = 'size=' + imgStat + '; path=../';
}

function changeLayout(imgStat = imageLayout) {
  switch (imgStat) {
    case 0:
      imageLayout = 1;
      document.getElementById('layout').textContent = 'Layout all';
      Array.from(images).forEach((item) => (item.style.display = 'block'));
      Array.from(images)[getPageNum()].scrollIntoView();
      break;
    case 1:
      imageLayout = 0;
      document.getElementById('layout').textContent = 'Layout';
      Array.from(images).forEach((item) => (item.style.display = 'none'));
      Array.from(images)[getPageNum()].style.display = 'block';
      Array.from(images)[getPageNum()].scrollIntoView();
      break;
    default:
      break;
  }
  document.cookie = 'layout=' + imgStat + '; path=../';
}

const start = () => {
  startKeyHandler();
  images = Array.from(document.getElementsByClassName('ImageDisplay'))[0]
    .children[0]
    .children;
  pages = document.getElementById('pages');
  images[getPageNum()].style.display = 'block';

  const chapterIndex = parseInt(document.location.pathname.split('/')[3]);
  const pagesIndex = getPageNum();
  document.getElementById('chapters')[chapterIndex].selected = 'selected';
  document.getElementById('pages')[pagesIndex].selected = 'selected';

  const size = getCookie()['size'] | 0;
  const layout = getCookie()['layout'] | 0;
  changeImageSize(size);
  changeLayout(layout);

  document.cookie = 'lastVisited=' + document.location.pathname + '; path=../';
};
