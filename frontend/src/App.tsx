import React from "react";
import { Header } from "./components/Header";
import { History } from "./components/History";
import { Usage } from "./components/Usage";
import styles from './App.module.css';

export const App = () => (
  <>
    <Header />

    <div className={styles.Content}>
      <p className={styles.Para1}>KeyConjurer is an application for generating temporary session credentials for AWS and Tencent Cloud.</p>

      <div className={styles.History}>
        <History />
      </div>

      <div className={styles.Usage}>
        <Usage />
      </div>
    </div>
  </>
);
