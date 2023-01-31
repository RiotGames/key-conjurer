import React, { Component } from "react";
import { Form, Card, Message } from "semantic-ui-react";
import { authenticate, updateUserInfo } from "./../actions";
import { subscribe } from "./../stores";
import * as PropTypes from "prop-types";

class LoginForm extends Component {
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

  handleTextChange = (property) => (_event, data) => {
    const { username, password } = this.state;
    updateUserInfo({
      username: "username" === property ? data.value : username,
      password: "password" === property ? data.value : password,
    });
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

  authUser = (_event) => {
    const { username, password } = this.state;
    this.setState({
      loginRequest: true,
      error: false,
      errorMessage: "",
    });
    authenticate(username, password, this.props.idp);
  };

  handleSubmit = () => {
    this.authUser();
  };

  render() {
    const { username, password, requestSent, error, errorEvent, errorMessage } =
      this.state;

    return (
      <Card fluid>
        <Card.Content>
          <Card.Header>Auth</Card.Header>
          <Card.Description>
            <Form loading={requestSent} onSubmit={this.handleSubmit}>
              <Form.Group widths="equal" inline>
                <Form.Input
                  fluid
                  label="Username"
                  placeholder="Username"
                  value={username}
                  onChange={this.handleTextChange("username")}
                  autoFocus
                  autoComplete="off"
                />
                <Form.Input
                  fluid
                  type="password"
                  label="Password"
                  placeholder="Password"
                  value={password}
                  onChange={this.handleTextChange("password")}
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

export default LoginForm;

LoginForm.propTypes = {
  idp: PropTypes.string.isRequired,
};
