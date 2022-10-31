import yaml from "yaml";

export function isObject(item) {
  return item && typeof item === "object" && !Array.isArray(item);
}

export function mergeDeep(target, ...sources) {
  if (!sources.length) return target;
  const source = sources.shift();

  if (isObject(target) && isObject(source)) {
    for (const key in source) {
      if (isObject(source[key])) {
        if (!target[key]) Object.assign(target, { [key]: {} });
        mergeDeep(target[key], source[key]);
      } else {
        Object.assign(target, { [key]: source[key] });
      }
    }
  }

  return mergeDeep(target, ...sources);
}

export function sha256(ascii) {
  function rightRotate(value, amount) {
    return (value >>> amount) | (value << (32 - amount));
  }

  var mathPow = Math.pow;
  var maxWord = mathPow(2, 32);
  var lengthProperty = "length";
  var i, j;
  var result = "";

  var words = [];
  var asciiBitLength = ascii[lengthProperty] * 8;

  var hash = (sha256.h = sha256.h || []);
  var k = (sha256.k = sha256.k || []);
  var primeCounter = k[lengthProperty];
  var isComposite = {};
  for (var candidate = 2; primeCounter < 64; candidate++) {
    if (!isComposite[candidate]) {
      for (i = 0; i < 313; i += candidate) {
        isComposite[i] = candidate;
      }
      hash[primeCounter] = (mathPow(candidate, 0.5) * maxWord) | 0;
      k[primeCounter++] = (mathPow(candidate, 1 / 3) * maxWord) | 0;
    }
  }

  ascii += "\x80";
  while ((ascii[lengthProperty] % 64) - 56) ascii += "\x00";
  for (i = 0; i < ascii[lengthProperty]; i++) {
    j = ascii.charCodeAt(i);
    if (j >> 8) return;
    words[i >> 2] |= j << (((3 - i) % 4) * 8);
  }
  words[words[lengthProperty]] = (asciiBitLength / maxWord) | 0;
  words[words[lengthProperty]] = asciiBitLength;

  for (j = 0; j < words[lengthProperty]; ) {
    var w = words.slice(j, (j += 16));
    var oldHash = hash;
    hash = hash.slice(0, 8);

    for (i = 0; i < 64; i++) {
      var w15 = w[i - 15],
        w2 = w[i - 2];

      var a = hash[0],
        e = hash[4];
      var temp1 =
        hash[7] +
        (rightRotate(e, 6) ^ rightRotate(e, 11) ^ rightRotate(e, 25)) +
        ((e & hash[5]) ^ (~e & hash[6])) +
        k[i] +
        (w[i] =
          i < 16
            ? w[i]
            : (w[i - 16] +
                (rightRotate(w15, 7) ^ rightRotate(w15, 18) ^ (w15 >>> 3)) +
                w[i - 7] +
                (rightRotate(w2, 17) ^ rightRotate(w2, 19) ^ (w2 >>> 10))) |
              0);
      var temp2 =
        (rightRotate(a, 2) ^ rightRotate(a, 13) ^ rightRotate(a, 22)) +
        ((a & hash[1]) ^ (a & hash[2]) ^ (hash[1] & hash[2]));

      hash = [(temp1 + temp2) | 0].concat(hash);
      hash[4] = (hash[4] + temp1) | 0;
    }

    for (i = 0; i < 8; i++) {
      hash[i] = (hash[i] + oldHash[i]) | 0;
    }
  }

  for (i = 0; i < 8; i++) {
    for (j = 3; j + 1; j--) {
      var b = (hash[i] >> (j * 8)) & 255;
      result += (b < 16 ? 0 : "") + b.toString(16);
    }
  }
  return result;
}

export function filterArrayIncludes(array, { key, value }) {
  if (!array || !Array.isArray(array) || !array.length) {
    return [];
  }

  return array.filter(
    (item) =>
      (typeof item[key] === "string" &&
        item[key].toLowerCase().startsWith(value)) ||
      false
  );
}

export function filterArrayBy(array, { key, value }) {
  if (!array || !Array.isArray(array) || !array.length) {
    return [];
  }

  return array.filter((item) => item[key] === value);
}

export function filterArrayByTitleAndUuid(
  array,
  value,
  unique = true,
  titleKey = "title"
) {
  const byUuid = filterArrayIncludes(array, {
    key: "uuid",
    value: value.toLowerCase(),
  });

  const byTitle = filterArrayIncludes(array, {
    key: titleKey,
    value: value.toLowerCase(),
  });

  if (!unique) {
    return byTitle.concat(byUuid);
  }

  return [...new Set([...byTitle, ...byUuid])];
}

export function levenshtein(s, t) {
  const d = [];

  const n = s.length,
    m = t.length;

  if (n == 0) return m;
  if (m == 0) return n;

  for (let ik = n; ik >= 0; ik--) d[ik] = [];

  for (let ix = n; ix >= 0; ix--) d[ix][0] = i;
  for (let jf = m; jf >= 0; jf--) d[0][jf] = jf;

  for (var i = 1; i <= n; i++) {
    const s_i = s.charAt(i - 1);

    for (let j = 1; j <= m; j++) {
      if (i == j && d[i][j] > 4) return n;

      const t_j = t.charAt(j - 1);
      const cost = s_i == t_j ? 0 : 1;

      let mi = d[i - 1][j] + 1;
      const b = d[i][j - 1] + 1;
      const c = d[i - 1][j - 1] + cost;

      if (b < mi) mi = b;
      if (c < mi) mi = c;

      d[i][j] = mi;

      if (i > 1 && j > 1 && s_i == t.charAt(j - 2) && s.charAt(i - 2) == t_j) {
        d[i][j] = Math.min(d[i][j], d[i - 2][j - 2] + cost);
      }
    }
  }

  return d[n][m];
}

export function downloadFile(blob, name, extension = "json") {
  if (window.navigator.msSaveOrOpenBlob) {
    window.navigator.msSaveBlob(blob, name);
  } else {
    const elem = window.document.createElement("a");
    elem.href = window.URL.createObjectURL(blob);
    elem.download = name + "." + extension;
    document.body.appendChild(elem);
    elem.click();
    document.body.removeChild(elem);
  }
}

export function downloadJSONFile(obj, name) {
  const blob = new Blob([JSON.stringify(obj)], {
    type: "application/json",
  });
  downloadFile(blob, name);
}

export function objectToYAMLString(obj) {
  const doc = new yaml.Document();
  doc.contents = obj;

  return doc.toString();
}

export function downloadYAMLFile(obj, name) {
  const blob = new Blob([objectToYAMLString(obj)], {});

  downloadFile(blob, name, "yaml");
}

export function readJSONFile(file) {
  return new Promise((resolve) => {
    if (!file) return;

    let reader = new FileReader();
    reader.onload = (e) => {
      const result = JSON.parse(e.target.result);
      resolve(result);
    };
    reader.readAsText(file);
  });
}

export function readYAMLFile(file) {
  return new Promise((resolve) => {
    if (!file) return;

    let reader = new FileReader();
    reader.onload = (e) => {
      const result = yaml.parse(e.target.result);
      resolve(result);
    };
    reader.readAsText(file);
  });
}
