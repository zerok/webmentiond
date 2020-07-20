const Path = require('path');
const VueLoaderPlugin = require('vue-loader/lib/plugin');

module.exports = {
  entry: {
      bundle: './index.js', 
      widget: './widget.js'
  },
  mode: 'development',
  module: {
    rules: [
      {
        test: /\.vue$/,
        loader: 'vue-loader'
      }, {
        test: /\.svg$/,
        loader: 'file-loader',
        options: {
          esModule: false,
          publicPath: 'dist'
        }
      }
    ]
  },
  output: {
    path: Path.resolve(__dirname, 'dist'),
    filename: '[name].js'
  },
  plugins: [
    new VueLoaderPlugin()
  ],
  watchOptions: {
    ignored: /node_modules/,
    poll: 1000
  }
};
