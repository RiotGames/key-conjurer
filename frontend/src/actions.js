import { update, save, resetStores } from "./stores";
import { keyConjurerApiUrl, client } from "./consts";
import { version as clientVersion } from "./version";

export async function authenticate(username, password, idp) {
  resetStores(["idpInfo", "awsKeys"]);
  update("request", { requestSent: true });

  try {
    const response = await fetch(`${keyConjurerApiUrl}/get_user_data`, {
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
        authentication_provider: idp,
      }),
    });

    const { Success, Message, Data } = await response.json();
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

    const apps = Data.apps.map((app) => {
      return {
        id: app["@id"],
        name: app.name,
      };
    });

    update("idpInfo", { apps });
  } catch (error) {
    update("errors", {
      message: error.message,
      error: true,
      event: "login",
    });
  } finally {
    update("request", { requestSent: false });
  }
}

export async function requestKeys({
  username,
  password,
  selectedAccount,
  timeout,
  idp,
  role,
}) {
  resetStores(["awsKeys"]);
  update("request", { requestSent: true });

  const body = {
    username,
    password,
    client,
    clientVersion,
    appId: selectedAccount,
    timeoutInHours: timeout,
    authentication_provider: idp,
    roleName: role,
  };

  // OneLogin does not require roles
  if (idp === "onelogin") {
    delete body.roleName;
  }

  try {
    const response = await fetch(`${keyConjurerApiUrl}/get_aws_creds`, {
      method: "POST",
      mode: "cors",
      headers: {
        "content-type": "application/json",
      },
      body: JSON.stringify(body),
    });

    const { Success, Message, Data } = await response.json();
    if (!Success) {
      throw Error(Message);
    }

    update("awsKeys", {
      accessKeyId: Data.AccessKeyId,
      secretAccessKey: Data.SecretAccessKey,
      sessionToken: Data.SessionToken,
      expiration: Data.Expiration,
    });
  } catch (error) {
    update("errors", {
      message: error.message,
      error: true,
      event: "keyRequest",
    });
  } finally {
    update("request", { requestSent: false });
  }
}

export function updateUserInfo({ username, password }) {
  update("userInfo", { username, password });
}
