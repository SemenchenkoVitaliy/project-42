const getCookie = () => {
  let result = {};

  document.cookie.split(';').forEach((cookie) => {
    cookie = cookie.trim();
    result[cookie.split('=')[0]] = cookie.split('=')[1];
  })
  return result
}

const load = () => {
  start();
  let path = getCookie()['lastVisited'];
  if (path !== undefined) {
    const lastPage = getCookie()['lastPage'];
    const hash = (lastPage === undefined) ? '' : ('#' + lastPage);
    document.getElementById('aLV').href = path + hash;
    document.getElementById('pLV').style.display = 'block';
  }
}
