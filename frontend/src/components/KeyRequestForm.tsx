import React, { Component, type ChangeEvent } from "react";
import { Message, Form, Card } from "semantic-ui-react";
import { requestKeys } from "../actions";
import { subscribe } from "../stores";

const documentationURL = process.env.REACT_APP_DOCUMENTATION_URL;
const timeouts = [1, 2, 3, 4, 5, 6, 7, 8];

interface RoleInputProps {
  onChange: (event: ChangeEvent<HTMLInputElement>) => void;
  value: string;
}

const RoleInput = ({ onChange, value }: RoleInputProps) => {
  return (
    <>
      <Form.Input
        fluid
        label="Role"
        onChange={onChange}
        value={value}
        list="common-roles"
      />
      <datalist id="common-roles">
        <option value="GL-Power">GL-Power</option>
        <option value="GL-NetEng">GL-NetEng</option>
        <option value="GL-Admin">GL-Admin</option>
      </datalist>
      <Message info>
        The suggestions provided are there for convenience, but may not
        necessarily match up to the roles you have access to. You can find out
        what roles you have access to by checking out{" "}
        <a href={documentationURL}>this confluence page</a>.
      </Message>
    </>
  );
};

interface State {
  accounts: { name: string; id: string }[];
  password: string;
  selectedAccount?: string;
  timeout: number;
  username: string;

  error: boolean;
  errorEvent: string;
  errorMessage?: string;
  role: string;

  keyRequest: boolean;
  requestSent: boolean;
}

export class KeyRequestForm extends Component<{}, State> {
  state: State = {
    accounts: [],
    keyRequest: false,
    password: "",
    requestSent: false,
    selectedAccount: undefined,
    timeout: parseInt(localStorage.getItem("timeout") ?? "1", 10),
    username: "",
    error: false,
    errorEvent: "",
    errorMessage: "",
    role: "",
  };

  subs: (() => void)[] = [];

  handleChange =
    <K extends keyof State>(name: K) =>
      (event: ChangeEvent<HTMLSelectElement | HTMLInputElement>) => {
        // We capture the value in a local value here to make sure event.target doesn't change underneath us
        const value = event.target.value;
        if (name === "timeout") {
          localStorage.setItem("timeout", value);
        }

        this.setState((prevState) => {
          return { ...prevState, [name]: value };
        });
      };

  setAccount = (event: ChangeEvent<HTMLSelectElement>) => {
    this.setState({ selectedAccount: event.target.value });
  };

  componentDidMount() {
    this.subs.push(
      subscribe("idpInfo", ({ apps: accounts }) => {
        // when idpInfo is triggered, we may have not selected an account yet, but
        // the DOM will visually indicate we have selected the first element.
        // The easiest way to fix this is to update the selected account to the first
        // available, if one has not already been selected.
        //
        // This doesn't get reproduced in tests because the tests can only test the
        // visual output to the browser, but this is happening within the internal state
        // of this component. Even when specifying that the browser should use the internal
        // state value, this still happens, which is very strange.
        this.setState((prevState) => {
          if (prevState.selectedAccount !== undefined) {
            // An account was selected by the user.
            if (
              accounts.length > 0 &&
              accounts.every((acc) => acc.id !== prevState.selectedAccount)
            ) {
              // There are no accounts in the new account set that match the users current account, so we should unset.
              // This is unlikely to ever happen - a user would have to be removed from an account between two button presses.
              return { accounts, selectedAccount: undefined };
            }

            // The only case left that a user had selected an account, and is still entitled to it.
            return { accounts, selectedAccount: prevState.selectedAccount };
          }

          if (accounts.length > 0 && prevState.accounts.length === 0) {
            // The user was not entitled to any accounts, and now they are - we should pick the first one.
            // This resolves #18.
            return { accounts, selectedAccount: accounts[0].id };
          }

          // The user had no account preference, and there are no special cases for the accounts list,
          // so preserve the current selection.
          return { accounts };
        });
      })
    );

    this.subs.push(
      subscribe("userInfo", ({ username, password }) => {
        this.setState({
          username,
          password,
        });
      })
    );

    this.subs.push(
      subscribe("request", ({ requestSent }) => {
        this.setState({
          requestSent,
          keyRequest: requestSent && this.state.keyRequest,
        });
      })
    );
    this.subs.push(
      subscribe("errors", ({ error, event, message: errorMessage }) => {
        if (event === "keyRequest") {
          this.setState({ error, errorMessage, errorEvent: event });
        }
      })
    );
  }

  componentWillUnmount(): void {
    for (const unsub of this.subs) {
      unsub();
    }
  }

  handleSubmit = (event: unknown) => {
    const { username, password, selectedAccount, timeout, role } = this.state;
    this.setState({
      keyRequest: true,
      error: false,
      errorEvent: "",
      errorMessage: "",
    });

    requestKeys({
      username,
      password,
      selectedAccount: selectedAccount!,
      timeout,
      role,
    });
  };

  render() {
    const {
      accounts,
      keyRequest,
      requestSent,
      timeout,
      error,
      errorEvent,
      errorMessage,
      role,
      selectedAccount,
    } = this.state;

    const accountPlaceHolder = accounts.length ? "Accounts" : "No Accounts";
    // Dropdowns use raw HTML with Semantic UI React classes to make them accessible.
    // Elements must be accessible to be accessed from React Testing Library in tests
    // (and also to obey California law).
    //
    // Semantic UI React uses divs for everything which cannot be labeled and is thus
    // not accessible.
    return (
      <Card fluid>
        <Card.Content>
          <Card.Header>Key Request</Card.Header>
          <Card.Meta>Will MFA push on request</Card.Meta>
          <Card.Description>
            <Form loading={requestSent} onSubmit={this.handleSubmit}>
              <Form.Group widths="equal">
                <Form.Field>
                  <label htmlFor="account">Account</label>
                  <select
                    className="ui field dropdown selection"
                    disabled={!accounts.length}
                    id="account"
                    placeholder={accountPlaceHolder}
                    onChange={this.setAccount}
                    value={selectedAccount}
                  >
                    {!selectedAccount && (
                      <option hidden>{accountPlaceHolder}</option>
                    )}

                    {accounts.map(({ name, id }) => (
                      <option key={id} value={id}>
                        {name}
                      </option>
                    ))}
                  </select>
                </Form.Field>

                <Form.Field>
                  <label htmlFor="timeout">TTL (hours)</label>
                  <select
                    className="ui field dropdown selection"
                    id="timeout"
                    onChange={this.handleChange("timeout")}
                    value={timeout}
                  >
                    {timeouts.map((timeout) => (
                      <option key={timeout} value={timeout}>
                        {timeout}
                      </option>
                    ))}
                  </select>
                </Form.Field>
              </Form.Group>
              <RoleInput value={role} onChange={this.handleChange("role")} />
              <Form.Button fluid primary type="submit">
                Request Keys
              </Form.Button>
            </Form>
            {keyRequest ? <Message positive>Sending Push</Message> : ""}
            {error && errorEvent === "keyRequest" ? (
              <Message negative>{errorMessage}</Message>
            ) : (
              ""
            )}
          </Card.Description>
        </Card.Content>
      </Card>
    );
  }
}
