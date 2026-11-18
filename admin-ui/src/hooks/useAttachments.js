import api from "@/api.js";

// Files are stored by the core-chatting attachments service (S3):
//   PUT  /attachments        registers a file, returns its uuid and a
//                            presigned S3 url to upload the content to
//   GET  /attachments?ids=   { [uuid]: { url, title } }, url without scheme
// The S3 links are public: anyone with the link can open the file.

// core-chatting builds its JSON responses by string concatenation, so quotes
// and backslashes in a file name would break them
const safeTitle = (name) => name.replace(/["\\]/g, "_");

const toHttps = (url) => (/^https?:\/\//.test(url) ? url : `https://${url}`);

const cache = new Map(); // uuid -> { url, title }

/** Upload a file, returns { uuid, url, title } */
export const uploadAttachment = async (file) => {
  const title = safeTitle(file.name);
  const { data } = await api.axios.put("attachments", { title, chat: "" });
  const { uuid, object_id: uploadUrl } = data;

  // a plain request: the presigned url must not get our auth headers
  const response = await fetch(uploadUrl, { method: "PUT", body: file });
  if (!response.ok) {
    throw new Error(`Upload of ${file.name} failed: ${response.status}`);
  }

  const [attachment] = Object.values(await getAttachments([uuid]));
  return attachment ?? { uuid, title, url: "" };
};

/** { [uuid]: { uuid, url, title } } for the given uuids (missing are skipped) */
export const getAttachments = async (uuids) => {
  const missing = [...new Set(uuids)].filter((uuid) => uuid && !cache.has(uuid));

  if (missing.length) {
    const { data } = await api.axios.get("attachments", {
      params: { ids: missing.join(",") },
    });
    Object.entries(data ?? {}).forEach(([uuid, { url, title }]) => {
      cache.set(uuid, { uuid, url: toHttps(url), title });
    });
  }

  return Object.fromEntries(
    uuids.filter((uuid) => cache.has(uuid)).map((uuid) => [uuid, cache.get(uuid)]),
  );
};

export const isImage = (title = "") =>
  /\.(png|jpe?g|gif|webp|bmp|svg)$/i.test(title);
