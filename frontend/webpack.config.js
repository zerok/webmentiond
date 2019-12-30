const Path = require('path');
const VueLoaderPlugin = require('vue-loader/lib/plugin');

module.exports = {
  entry: './index.js',
  mode: 'development',
  module: {
    rules: [
      {
        test: /\.vue$/,
        loader: 'vue-loader'
      }
    ]
  },
  output: {
    path: Path.resolve(__dirname, 'dist'),
    filename: 'bundle.js'
  },
  plugins: [
    new VueLoaderPlugin()
  ]
};
