import React from "react";
import { Menu } from "semantic-ui-react";

export const Header = () => {
  return (
    <Menu fixed="top" fluid color="grey">
      <Menu.Item header>Key Conjurer</Menu.Item>
      <Menu.Item position="right">{process.env.REACT_APP_VERSION}</Menu.Item>
    </Menu>
  );
};
