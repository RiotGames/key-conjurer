import { update, save, resetStores } from "./stores";

import { keyConjurerApiUrl, client } from "./consts";
import { version as clientVersion } from "./version";

export function authenticate(username, password, shouldEncryptCreds) {
  resetStores(["idpInfo", "awsKeys"]);
  update("request", { requestSent: true });

  fetch(`${keyConjurerApiUrl}/get_user_data`, {
    method: "POST",
    mode: "cors",
    headers: {
      "content-type": "application/json",
    },
    body: JSON.stringify({
      username,
      password,
      client,
      clientVersion,
      shouldEncryptCreds: shouldEncryptCreds,
    }),
  })
    .then((res) => res.json())
    .then(({ Success, Message, Data }) => {
      if (!Success) {
        throw Error(Message);
      }
      if (Data.creds) {
        save("creds", Data.creds);
        update("userInfo", {
          username: "encrypted",
          password: Data.creds,
        });
      }
      update("idpInfo", { apps: Data.apps });
      update("request", { requestSent: false });
    })
    .catch((error) => {
      update("request", { requestSent: false });
      update("errors", {
        message: error.message,
        error: true,
        event: "login",
      });
    });
}

export function requestKeys({ username, password, selectedAccount, timeout }) {
  resetStores(["awsKeys"]);
  update("request", { requestSent: true });

  fetch(`${keyConjurerApiUrl}/get_aws_creds`, {
    method: "POST",
    mode: "cors",
    headers: {
      "content-type": "application/json",
    },
    body: JSON.stringify({
      username,
      password,
      client,
      clientVersion,
      appId: `${selectedAccount}`,
      timeoutInHours: timeout,
    }),
  })
    .then((res) => res.json())
    .then(({ Success, Message, Data }) => {
      if (!Success) {
        throw Error(Message);
      }
      update("awsKeys", {
        accessKeyId: Data.accessKeyId,
        secretAccessKey: Data.secretAccessKey,
        sessionToken: Data.sessionToken,
        expiration: Data.expiration,
      });
      update("request", { requestSent: false });
    })
    .catch((error) => {
      update("request", { requestSent: false });
      update("errors", {
        message: error.message,
        error: true,
        event: "keyRequest",
      });
    });
}

export function updateUserInfo({ username, password }) {
  update("userInfo", { username, password });
}
