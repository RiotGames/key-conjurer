import React, { Component } from 'react';
import { Menu } from 'semantic-ui-react';
import { version } from '../version';

class Header extends Component {
    render() {
	return (
            <Menu fixed='top' fluid color='grey'>
              <Menu.Item header>Key Conjurer</Menu.Item>
              <Menu.Item position='right'>{version}</Menu.Item>
            </Menu>
	);
    }
}

export default Header;
