import { keys, clone } from "ramda";
import { EventEmitter } from "events";

const events = new EventEmitter();

const defaultStores = {
  userInfo: {
    username: "",
    password: "",
  },
  idpInfo: {
    apps: [],
  },
  awsKeys: {
    accessKeyId: "",
    secretAccessKey: "",
    sessionToken: "",
    expiration: "",
    requestingKeys: false,
  },
  request: {
    requestSent: false,
  },
  errors: {
    error: false,
    event: "",
  },
};

const stores = clone(defaultStores);

export function save(key, value) {
  localStorage[key] = value;
}

export function resetStores(stores) {
  stores.map((store) => resetStore(store));
}

export function resetAllStores() {
  resetStores(keys(stores));
}

export function update(store, value) {
  if (stores[store]) {
    stores[store] = { ...stores[store], ...value };
    events.emit(`${store}Updated`, stores[store]);
  } else {
    console.log(`No store named ${store}`);
  }
}

export function subscribe(store, cb) {
  stores[store]
    ? events.on(`${store}Updated`, (_) => cb(stores[store]))
    : console.log(`No store named ${store}`);
}

function resetStore(store) {
  if (stores[store]) {
    stores[store] = clone(defaultStores[store]);
    events.emit(`${store}Updated`, stores[store]);
  } else {
    console.log(`No store named ${store}`);
  }
}
