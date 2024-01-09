import { utils, write } from "xlsx";
import { saveAs } from "file-saver";

function s2ab(s) {
  var buf = new ArrayBuffer(s.length); //convert s to arrayBuffer
  var view = new Uint8Array(buf); //create uint8array as viewer
  for (var i = 0; i < s.length; i++) view[i] = s.charCodeAt(i) & 0xff; //convert to octet
  return buf;
}

export default class XlsxService {
  static downloadXlsx(name, pages) {
    const wb = this.createBook(name);

    pages.forEach((page) => {
      const headers = page.headers;
      const items = [];
      page.items.forEach((item, index) => {
        Object.keys(item).forEach((key) => {
          if (!items[index]) {
            items[index] = [];
          }
          let subIndex = headers.findIndex((h) => h.key === key);

          items[index][subIndex] = item[key];
        });
      });

      this.generatePageByHeaderAndItems(wb, {
        items: items,
        headers,
        name: page.name,
      });
    });

    return this.downloadSheets(name, wb);
  }

  static createBook(name) {
    var wb = utils.book_new();
    wb.props = {
      title: name,
      subject: name,
      author: "Nocloud",
      createdDate: Date.now(),
    };
    return wb;
  }

  static addTosheet(wb, name, mat) {
    name = name.slice(0, 25);

    wb.SheetNames.push(name);
    var ws_data = JSON.parse(JSON.stringify(mat));
    var ws = utils.aoa_to_sheet(ws_data);
    wb.Sheets[name] = ws;
  }

  static download(name, wbout) {
    return saveAs(
      new Blob([s2ab(wbout)], { type: "application/octet-stream" }),
      name + ".xlsx"
    );
  }

  static generatePageByHeaderAndItems(wb, { items, headers, name }) {
    const mat = [];

    mat[0] = headers.map((h) => h.title);
    items.forEach((item) => {
      mat.push(item);
    });

    this.addTosheet(wb, name, mat);
  }

  static downloadSheets(name, wb) {
    var wbout = write(wb, { bookType: "xlsx", type: "binary" });
    return this.download(name, wbout);
  }
}
