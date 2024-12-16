import yaml from "yaml";
import XlsxService from "@/services/XlsxService";
import store from "@/store";
import { Rounding } from "nocloud-proto/proto/es/billing/billing_pb";

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

export function filterArrayIncludes(array, { keys, value, params }) {
  if (!array || !Array.isArray(array) || !array.length) {
    return [];
  }

  return array.filter((item) =>
    keys.some((key) => {
      const newKey = params?.[key] ? params?.[key] : key;
      let newValue = item[newKey];

      switch (typeof params?.[key]) {
        case "function":
          newValue = newKey(item);
          break;
        case "string":
          newValue = item[key][newKey];
      }

      return (
        typeof newValue === "string" &&
        newValue.toLowerCase().includes(value.toLowerCase())
      );
    })
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
    keys: ["uuid"],
    value: value.toLowerCase(),
  });

  const byTitle = filterArrayIncludes(array, {
    keys: [titleKey],
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

export function getSecondsByDays(days) {
  return +days * 60 * 60 * 24;
}

export function getState(item) {
  if (!item.state) return item?.data?.is_monitored ? "ERROR" : "LCM_INIT";

  return item.state.state;
}

export function toKebabCase(str) {
  return str.replace(/([a-z])([A-Z])/g, "$1-$2").toLowerCase();
}

export function toPascalCase(text) {
  if (!text) {
    return;
  }
  return text.replace(/(^\w|-\w)/g, (text) =>
    text.replace(/-/, "").toUpperCase()
  );
}

export function formatSecondsToDate(timestamp, withTime, sep = ".") {
  if (!timestamp || !Number(timestamp)) return;
  const date = new Date(Number(timestamp) * 1000);
  const time = date
    .toLocaleString(undefined, {
      hourCycle: "h24",
      timeZone: store.getters["settings/timeZone"],
    })
    .split(" ")[1];

  const year = date.toUTCString().split(" ")[3];
  let month = date.getUTCMonth() + 1;
  let day = date.getUTCDate();

  if (`${month}`.length < 2) month = `0${month}`;
  if (`${day}`.length < 2) day = `0${day}`;

  let result = `${day}${sep}${month}${sep}${year}`;

  if (withTime) result += ` ${time}`;
  return result;
}

export function formatDateToTimestamp(strDate) {
  const datum = Date.parse(strDate);
  return datum / 1000;
}

export function formatSecondsToDateString(timestamp) {
  if (!timestamp || !+timestamp) return;
  const date = new Date(timestamp * 1000);

  const year = date.toUTCString().split(" ")[3];
  let month = date.getUTCMonth() + 1;
  let day = date.getUTCDate();

  if (`${month}`.length < 2) month = `0${month}`;
  if (`${day}`.length < 2) day = `0${day}`;

  let result = `${year}-${month}-${day}`;

  return result;
}

export function getTimestamp({ day, month, year, quarter, week, time }) {
  let seconds = 0;

  seconds += getSecondsByDays(30 * month);
  seconds += getSecondsByDays(30 * 3 * quarter);
  seconds += getSecondsByDays(7 * week);
  seconds += getSecondsByDays(365 * year);
  seconds += getSecondsByDays(day);
  seconds +=
    new Date("1970-01-01T" + time + "Z").getTime() / 1000
      ? new Date("1970-01-01T" + time + "Z").getTime() / 1000
      : 0;

  return seconds;
}

export function getOvhPrice(instance) {
  const duration = instance.config.duration;
  const tarrifPrice =
    instance.billingPlan.products[`${duration} ${instance.config.planCode}`]
      ?.price;
  const addonsPrice = instance.config.addons
    ?.map(
      (a) =>
        instance.billingPlan.resources.find((r) => r.key === `${duration} ${a}`)
          ?.price || 0
    )
    .reduce((acc, v) => acc + v, 0);
  return tarrifPrice + addonsPrice;
}

export function getFullDate(period) {
  period = +period;
  const dayEqualent = 86400;
  const result = {
    day: 0,
    year: 0,
    month: 0,
    quarter: 0,
    week: 0,
    time: "",
  };

  result.day = Math.floor(period / dayEqualent);
  period -= result.day * dayEqualent;

  if (result.day === 90) {
    result.day = 0;
    result.quarter = 1;
    return result;
  }

  result.year = Math.floor(result.day / 365);
  result.day -= result.year * 365;

  result.month = Math.floor(result.day / 30);
  result.day -= result.month * 30;

  result.week = Math.floor(result.day / 7);
  result.day -= result.week * 7;

  result.time = new Date(period * 1000).toUTCString().split(" ").at(-2);

  return result;
}

export function getTodayFullDate() {
  const date = new Date();
  return (
    ("00" + (date.getMonth() + 1)).slice(-2) +
    "/" +
    ("00" + date.getDate()).slice(-2) +
    "/" +
    date.getFullYear() +
    " " +
    ("00" + date.getHours()).slice(-2) +
    ":" +
    ("00" + date.getMinutes()).slice(-2) +
    ":" +
    ("00" + date.getSeconds()).slice(-2)
  ).replace(" ", "-");
}

export function getMarginedValue(fee, val) {
  const n = Math.pow(10, fee.precision ?? 0);
  let percent = (fee?.default ?? 0) / 100 + 1;
  let round;

  switch (fee.round) {
    case 1:
      round = "floor";
      break;
    case 2:
      round = "round";
      break;
    case 3:
      round = "ceil";
      break;
    default:
      round = "round";
  }
  if (fee.round === "NONE" || !fee.round) round = "round";
  else if (typeof fee.round === "string") {
    round = fee.round.toLowerCase();
  }

  for (let range of fee?.ranges ?? []) {
    if (val <= range.from) continue;
    if (val > range.to) continue;
    percent = range.factor / 100 + 1;
  }

  return Math[round](val * percent * n) / n || 0;
}

export async function getClientIP() {
  const regexp = /[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}/;
  const response = await fetch("https://www.cloudflare.com/cdn-cgi/trace");
  const text = await response.text();
  return text.match(regexp)[0];
}

export function defaultFilterObject(item, queryText) {
  return (
    item?.title?.toLocaleLowerCase().indexOf(queryText.toLocaleLowerCase()) !==
      -1 ||
    item?.uuid?.toLocaleLowerCase().startsWith(queryText.toLocaleLowerCase())
  );
}

function fetchMDIIconsHash() {
  const icons = [];
  let block = null;

  return () => {
    if (block) {
      return block;
    }
    if (icons.length) {
      return icons;
    }
    block = fetch(
      "https://raw.githubusercontent.com/Templarian/MaterialDesign/master/meta.json",
      { method: "get" }
    ).then((d) => d.json());

    return block;
  };
}

export const fetchMDIIcons = fetchMDIIconsHash();

export function getBillingPeriod(period) {
  period = Number(period);

  if (period === 0) {
    return "One time";
  }

  const fullPeriod = period && getFullDate(period);
  if (!fullPeriod) {
    return {};
  }
  fullPeriod.hours = +fullPeriod.time.split(":")?.[0];
  fullPeriod.time = undefined;
  if (fullPeriod) {
    const period = Object.keys(fullPeriod)
      .filter((key) => +fullPeriod[key])
      .map((key) => `${fullPeriod[key]} ${key}s`)
      .join(", ");
    const beautifulAnnotations = {
      "1 months": "Monthly",
      "1 days": "Daily",
      "1 years": "Yearly",
      "1 quarters": "Quarter",
      "1 hourss": "Hourly",
    };
    return beautifulAnnotations[period] || period;
  }
}

export function filterByKeysAndParam(items, keys, param) {
  const searchParam = param?.toLowerCase();
  return items.filter((a) => {
    return keys.some((key) => {
      let data = a;
      key.split(".").forEach((subkey) => {
        data = data?.[subkey];
      });
      return (
        data &&
        data.toString().toLowerCase().includes(searchParam.toLowerCase())
      );
    });
  });
}

export function compareSearchValue(data, searchValue, field) {
  const type = field?.type || "";
  switch (type) {
    case "input": {
      return (
        !searchValue ||
        data?.toString().toLowerCase().includes(searchValue.toLowerCase())
      );
    }
    case "select": {
      return !searchValue.length || searchValue.includes(data);
    }
    case "logic-select": {
      return (
        (!searchValue && searchValue !== false) ||
        searchValue?.toString()?.toLowerCase() ===
          data?.toString()?.toLowerCase()
      );
    }
    case "number-range": {
      const isNegative = data?.toString().startsWith("-");
      data = +data?.toString().replace(/[^\d\\.]*/g, "");
      if (isNegative) {
        data = -data;
      }
      return (
        ((!searchValue?.from && searchValue?.from !== 0) ||
          +searchValue.from <= data) &&
        ((!searchValue?.to && searchValue?.to !== 0) || +searchValue.to >= data)
      );
    }
    case "date": {
      if (!searchValue) {
        return true;
      }

      if (typeof data === "number") data = new Date(data * 1000);
      data = new Date(data).getTime();
      const [first, second] = searchValue;
      if (first && second) {
        const min = (
          new Date(first).getTime() > new Date(second).getTime()
            ? new Date(second)
            : new Date(first)
        )?.getTime();
        const max = (
          new Date(first).getTime() < new Date(second).getTime()
            ? new Date(second)
            : new Date(first)
        )?.getTime();

        return data && min <= data && max >= data;
      } else {
        return data && data === new Date(first).toLocaleString();
      }
    }
    default: {
      return false;
    }
  }
}

export function getDeepObjectValue(data, key) {
  let value = { ...data };
  key.split(".").forEach((subKey, index) => {
    if (index === key.split(".").length - 1) {
      key = subKey;
      return;
    }
    value = data?.[subKey];
  });
  return value?.[key];
}

export function debounce(callback, wait = 100) {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      callback(...args);
    }, wait);
  };
}

export function addToClipboard(text) {
  if (navigator?.clipboard) {
    navigator.clipboard.writeText(text).catch((res) => {
      console.error(res);
    });
  } else {
    alert("Clipboard is not supported!");
  }
}

export function downloadPlanXlsx(plans) {
  const baseHeaders = [
    { title: "Title", key: "title" },
    { title: "Price", key: "price" },
    { title: "Period", key: "period" },
    { title: "Kind", key: "kind" },
    { title: "Group", key: "group" },
  ];

  return XlsxService.downloadXlsx(
    "Plans " + getTodayFullDate(),
    plans.map((plan) => {
      const headers = [...baseHeaders];
      switch (plan.type) {
        case "ovh vps": {
          headers.push({ title: "Base price", key: "basePrice" });
          headers.push({ title: "API name", key: "apiName" });
          headers.push({ title: "Region", key: "datacenter" });
          break;
        }
        case "ovh dedicated": {
          headers.push({ title: "Base price", key: "basePrice" });
          headers.push({ title: "API name", key: "apiName" });
          headers.push({ title: "CPU", key: "cpu" });
          headers.push({ title: "Region", key: "datacenter" });
          break;
        }
        case "ovh cloud": {
          headers.push({ title: "Region", key: "datacenter" });
          break;
        }
      }

      let configPrice = 0;

      if (plan.type === "ione" && plan.meta.minDiskSize) {
        const ip =
          plan.resources.find((r) => r.key === "ips_public")?.price || 0;

        configPrice += ip;

        const hddMinSize = +plan.meta.minDiskSize.HDD || 20;
        const ssdMinSize = +plan.meta.minDiskSize.SSD || 20;

        const hddMin =
          plan.resources.find((r) => r.key === "drive_hdd")?.price *
            hddMinSize || 0;
        const ssdMin =
          plan.resources.find((r) => r.key === "drive_ssd")?.price *
            ssdMinSize || 0;

        if (!ssdMin && hddMin) {
          configPrice += hddMin;
        }
        if (!hddMin && ssdMin) {
          configPrice += ssdMin;
        } else {
          configPrice += Math.min(ssdMin, hddMin);
        }
      }

      return {
        name: plan.title,
        headers,
        items: Object.values(plan.products)
          .map((product) => {
            const result = {};
            product = {
              ...product,
              ...product.meta,
              price: product.price + configPrice,
            };

            Object.keys(product).forEach((key) => {
              if (headers.findIndex((a) => a.key === key) !== -1) {
                result[key] = product[key];
              }
            });

            return result;
          })
          .map((p) => ({ ...p, period: getBillingPeriod(p.period) })),
      };
    })
  );
}

export function isInstancePayg(inst) {
  return (
    (inst.type === "ione" && inst.billingPlan.kind === "DYNAMIC") ||
    inst.type === "openai"
  );
}

export function getShortName(name = "", maxLength = 30) {
  return name.length > maxLength + 3
    ? name.slice(0, maxLength - 3) + "..."
    : name;
}

export function formatPrice(price, { precision, rounding } = {}) {
  price = +price || 0;
  precision = precision || 0;
  rounding = rounding || "ROUND_HALF";

  if (price < 0.01 && price > -1) {
    return parseFloat(price.toFixed(10));
  }

  if (price == 0) {
    return 0;
  }

  if (Rounding.ROUND_HALF === Rounding[rounding]) {
    return price.toFixed(precision).toString();
  }

  const fn =
    Rounding[rounding] === Rounding.ROUND_DOWN ? Math.floor : Math.round;

  return fn(price * Math.pow(10, precision)) / Math.pow(10, precision);
}
