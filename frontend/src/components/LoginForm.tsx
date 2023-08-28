import React, { Component } from "react";
import { Form, Card, Message } from "semantic-ui-react";
import { authenticate, updateUserInfo } from "../actions";
import { subscribe } from "../stores";

interface State {
  username: string;
  password: string;

  requestSent: boolean;
  loginRequest: boolean;
  encryptCreds: boolean;

  error: boolean;
  errorEvent: string;
  errorMessage?: string;
}

export class LoginForm extends Component<{}, State> {
  state = {
    username: "",
    password: "",
    requestSent: false,
    loginRequest: false,
    encryptCreds: false,
    error: false,
    errorEvent: "",
    errorMessage: "",
  };

  componentDidMount() {
    if (localStorage.creds) {
      updateUserInfo({
        username: "encrypted",
        password: localStorage.creds,
      });

      this.setState({
        username: "encrypted",
        password: localStorage.creds,
      });
    }

    subscribe("userInfo", ({ username, password }) => {
      this.setState({
        username,
        password,
      });
    });

    subscribe("request", ({ requestSent }) =>
      this.setState({
        requestSent,
        loginRequest: requestSent && this.state.loginRequest,
      })
    );

    subscribe("errors", ({ error, event, message: errorMessage }) => {
      if (event === "login") {
        this.setState({ error, errorMessage, errorEvent: event });
      }
    });
  }

  render() {
    const { username, password, requestSent, error, errorEvent, errorMessage } =
      this.state;

    const handleTextChange =
      (property: "username" | "password") =>
      (_event: unknown, data: { value: string }) => {
        const { username, password } = this.state;
        updateUserInfo({
          username: "username" === property ? data.value : username,
          password: "password" === property ? data.value : password,
        });
      };

    const handleSubmit = () => {
      const { username, password } = this.state;
      this.setState({
        loginRequest: true,
        error: false,
        errorMessage: "",
      });
      authenticate(username, password);
    };

    return (
      <Card fluid>
        <Card.Content>
          <Card.Header>Auth</Card.Header>
          <Card.Description>
            <Form loading={requestSent} onSubmit={handleSubmit}>
              <Form.Group widths="equal" inline>
                <Form.Input
                  fluid
                  label="Username"
                  placeholder="Username"
                  value={username}
                  onChange={handleTextChange("username")}
                  autoFocus
                  autoComplete="off"
                />
                <Form.Input
                  fluid
                  type="password"
                  label="Password"
                  placeholder="Password"
                  value={password}
                  onChange={handleTextChange("password")}
                  autoComplete="off"
                />
              </Form.Group>
              <Form.Button fluid primary type="submit">
                Authenticate
              </Form.Button>
            </Form>
          </Card.Description>
          {error && errorEvent === "login" && (
            <Message negative>{errorMessage}</Message>
          )}
        </Card.Content>
      </Card>
    );
  }
}
