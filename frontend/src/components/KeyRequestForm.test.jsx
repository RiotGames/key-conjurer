import * as React from "react";
import KeyRequestForm from "./keyRequestForm";
import { cleanup, render } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { update, resetAllStores } from "../stores";

beforeEach(() => {
	resetAllStores();
});

afterEach(cleanup);

test("should change account to the one selected", async () => {
	const user = userEvent.setup();
	const { getByLabelText } = render(<KeyRequestForm />);
	update("idpInfo", {
		apps: [
			{ name: "Account #1", id: 0 },
			{ name: "Account #2", id: 1 },
			{ name: "Account #3", id: 2 }
		]
	});

	const select = getByLabelText("Account");
	await user.selectOptions(select, "Account #2");

	expect(select.value).toEqual("1");
});

test("should change account to first account in list when the account list is changed", () => {
	const { getByLabelText } = render(<KeyRequestForm />);
	update("idpInfo", { apps: [] });

	const select = getByLabelText("Account");
	expect(select.value).toEqual("");

	update("idpInfo", {
		apps: [
			{ name: "Account #3", id: 2 },
			{ name: "Account #2", id: 1 },
			{ name: "Account #1", id: 0 },
		]
	});

	expect(select.value).toEqual("2");
});

test("should not change account when the accounts list changes if one has already been selected", async () => {
	const user = userEvent.setup();
	const { getByLabelText } = render(<KeyRequestForm />);
	update("idpInfo", {
		apps: [
			{ name: "Account #1", id: 0 },
			{ name: "Account #2", id: 1 },
			{ name: "Account #3", id: 2 },
		]
	});

	const select = getByLabelText("Account");
	await user.selectOptions(select, "Account #2");
	expect(select.value).toEqual("1");

	// Add a new value after the selected value
	update("idpInfo", {
		apps: [
			{ name: "Account #1", id: 0 },
			{ name: "Account #2", id: 1 },
			{ name: "Account #3", id: 2 },
			{ name: "Account #4", id: 3 },
		]
	});

	expect(select.value).toEqual("1");

	// Remove value from before the selected value
	update("idpInfo", {
		apps: [
			{ name: "Account #2", id: 1 },
			{ name: "Account #3", id: 2 },
			{ name: "Account #4", id: 3 },
		]
	});

	expect(select.value).toEqual("1");

	// Add a value before the selected value
	update("idpInfo", {
		apps: [
			{ name: "Account #1", id: 0 },
			{ name: "Account #2", id: 1 },
			{ name: "Account #3", id: 2 },
			{ name: "Account #4", id: 3 },
		]
	});

	expect(select.value).toEqual("1");
});

test("should change account to the first account in list if one has been selected when the account list changes, if the new account list does not include the currently selected account", async () => {
	const user = userEvent.setup();
	const { getByLabelText } = render(<KeyRequestForm />);
	update("idpInfo", {
		apps: [
			{ name: "Account #1", id: 0 },
			{ name: "Account #2", id: 1 },
			{ name: "Account #3", id: 2 },
		]
	});

	const select = getByLabelText("Account");
	await user.selectOptions(select, "Account #2");
	expect(select.value).toEqual("1");

	update("idpInfo", {
		apps: [
			{ name: "Account #1", id: 0 },
			{ name: "Account #3", id: 2 },
		]
	});

	expect(select.value).toEqual("0");
});
