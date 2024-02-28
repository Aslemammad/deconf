import assert from 'node:assert'
import {it} from 'node:test'
import { resolveConfig } from 'vite'

it('vite can read config', async () => {
  const config =await resolveConfig({}, "build") 
  assert.equal(config.base, '/fake-base/')
})
