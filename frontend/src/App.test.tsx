import React from "react";
import { App } from "./App";
import { test } from 'vitest';
import { render } from "@testing-library/react";

test("renders without crashing", () => {
  render(<App />);
});
