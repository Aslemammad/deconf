import assert from "node:assert";
import { it } from "node:test";
import { readFile } from "fs/promises";
import { resolveConfig } from "vite";

it("vite can read config", async () => {
  const config = await resolveConfig({}, "build");
  assert.equal(config.base, "/fake-base/");
});

it(".gitignore includes vite.config.ts", async () => {
  const content = await readFile("./.gitignore", "utf-8");
  assert.ok(content.includes("vite.config.ts"));
});
