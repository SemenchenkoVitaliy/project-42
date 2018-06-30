const getCookie = () => {
  const result = {};

  document.cookie.split(';').forEach((cookie) => {
    cookie = cookie.trim();
    result[cookie.split('=')[0]] = cookie.split('=')[1];
  });
  return result;
};

const load = () => {
  start();
  const path = getCookie()['lastVisited'];
  if (path !== undefined && document.getElementById('aLV')) {
    const lastPage = getCookie()['lastPage'];
    const hash = (lastPage === undefined) ? '' : ('#' + lastPage);
    document.getElementById('aLV').href = path + hash;
    document.getElementById('pLV').style.display = 'block';
  }
};
