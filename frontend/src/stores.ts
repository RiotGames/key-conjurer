import { keys, clone } from "ramda";
import { EventEmitter } from "events";

const events = new EventEmitter();

interface Application {
  id: string;
  name: string;
}

interface Store {
  userInfo: {
    username: string;
    password: string;
  };

  idpInfo: {
    apps: Application[];
  };

  awsKeys: {
    accessKeyId: string;
    secretAccessKey: string;
    sessionToken: string;
    expiration: string;
    requestingKeys: boolean;
  };

  request: {
    requestSent: boolean;
  };

  errors: {
    error: boolean;
    event: string;
    message: string | undefined;
  };
}

const defaultStores: Store = {
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
    message: undefined,
  },
};

const stores = clone(defaultStores);

export function resetStores<K extends keyof Store>(stores: K[]) {
  stores.map((store) => resetStore(store));
}

export function resetAllStores() {
  resetStores(keys(stores));
}

export function update<K extends keyof Store>(
  store: K,
  value: Partial<Store[K]>
) {
  stores[store] = { ...stores[store], ...value };
  events.emit(`${store}Updated`, stores[store]);
}

export function subscribe<K extends keyof Store>(
  store: K,
  cb: (store: Store[K]) => unknown
) {
  events.on(`${store}Updated`, (_) => cb(stores[store]));
}

function resetStore<K extends keyof Store>(store: K) {
  stores[store] = clone(defaultStores[store]);
  events.emit(`${store}Updated`, stores[store]);
}

export function save(key: string, value: string) {
  localStorage[key] = value;
}
