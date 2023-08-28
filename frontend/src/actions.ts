import { update, save, resetStores } from "./stores";

const apiURL = process.env.REACT_APP_API_URL;
const client = process.env.REACT_APP_CLIENT;
const clientVersion = process.env.REACT_APP_VERSION;

interface SuccessfulResponse<T> {
  Success: boolean;
  Message?: string;
  Data: T;
}

type Response<T> = SuccessfulResponse<T>;

interface ApplicationDto {
  "@id": string;
  name: string;
}

interface AuthenticateResponse {
  creds: string;

  apps: ApplicationDto[];
}

export async function authenticate(username: string, password: string) {
  resetStores(["idpInfo", "awsKeys"]);
  update("request", { requestSent: true });

  try {
    const response = await fetch(`${apiURL}/get_user_data`, {
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
        authentication_provider: "okta",
      }),
    });

    const { Success, Message, Data } =
      (await response.json()) as Response<AuthenticateResponse>;
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
  } catch (error: any) {
    update("errors", {
      message: error.message,
      error: true,
      event: "login",
    });
  } finally {
    update("request", { requestSent: false });
  }
}

interface KeyRequest {
  username: string;
  password: string;
  selectedAccount: string;
  timeout: number;
  role: string;
}

export async function requestKeys({
  username,
  password,
  selectedAccount,
  timeout,
  role,
}: KeyRequest) {
  resetStores(["awsKeys"]);
  update("request", { requestSent: true });

  const body = {
    username,
    password,
    client,
    clientVersion,
    appId: selectedAccount,
    timeoutInHours: timeout,
    authentication_provider: "okta",
    roleName: role,
  };

  try {
    const response = await fetch(`${apiURL}/get_aws_creds`, {
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
  } catch (error: any) {
    update("errors", {
      message: error.message,
      error: true,
      event: "keyRequest",
    });
  } finally {
    update("request", { requestSent: false });
  }
}

interface UserInfoRequest {
  username: string;
  password: string;
}

export function updateUserInfo({ username, password }: UserInfoRequest) {
  update("userInfo", { username, password });
}
