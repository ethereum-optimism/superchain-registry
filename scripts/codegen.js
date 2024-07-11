import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/**
 * Geneartes the addresses.json file that sits at the root of the
 * superchain/extra/addresses folder. Useful to have a combined JSON file for
 * all of the various addresses to avoid dynamic imports.
 */
const addresses = () => {
  const chainids = {}
  var folder = path.resolve(__dirname, '../superchain/configs')
  var subfolders = fs.readdirSync(folder, { withFileTypes: true })
  for (const subfolder of subfolders) {
    if (!subfolder.isDirectory()) {
      continue
    }

    const subpath = path.resolve(folder, subfolder.name)
    const files = fs.readdirSync(subpath)
    for (const file of files) {
      if (path.extname(file) !== '.toml') {
        continue
      }

      const filepath = path.resolve(subpath, file)
      const filename = filepath.replace('.toml', '')
      const chain = path.relative(folder, filename).replace(path.sep, '/')
      const content = fs.readFileSync(filepath, 'utf8')
      const matches = content.match(/^chain_id = \s*(\d+)$/m)

      if (matches) {
        const id = parseInt(matches[1])
        chainids[chain] = id
      }
    }
  }

  const result = {}
  folder = path.resolve(__dirname, '../superchain/extra/addresses')
  subfolders = fs.readdirSync(folder, { withFileTypes: true })
  for (const subfolder of subfolders) {
    if (!subfolder.isDirectory()) {
      continue
    }

    const subpath = path.resolve(folder, subfolder.name)
    const files = fs.readdirSync(subpath)
    for (const file of files) {
      if (path.extname(file) !== '.json') {
        continue
      }

      const filepath = path.resolve(subpath, file)
      const filename = filepath.replace('.json', '')
      const chain = path.relative(folder, filename).replace(path.sep, '/')
      const content = fs.readFileSync(filepath, 'utf8')

      if (chainids[chain]) {
        result[chainids[chain]] = JSON.parse(content)
      }
    }
  }

  const outpath = path.resolve(folder, 'addresses.json')
  fs.writeFileSync(outpath, JSON.stringify(result, null, 2))
}

addresses()
