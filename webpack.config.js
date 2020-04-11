const path = require('path');
const webpack = require('webpack');
const CopyWebpackPlugin = require('copy-webpack-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');

module.exports = {
  mode: 'development',
  // watch: true,
  watchOptions: {
    ignored: /node_modules/
  },
  target: 'node',
  context: __dirname + "/src",
  entry: './module.ts',
  output: {
    filename: "module.js",
    path: path.resolve(__dirname, 'dist'),
    libraryTarget: "amd"
  },
  externals: [
    'jquery', 'lodash', 'moment', 'react', 'angular',
    function(context, request, callback) {
      var prefix = 'grafana/';
      if (request.indexOf(prefix) === 0) {
        return callback(null, request.substr(prefix.length));
      }
      callback();
    }
  ],
  plugins: [
    new webpack.optimize.OccurrenceOrderPlugin(),
    new CopyWebpackPlugin([
      { from: 'plugin.json' },
      { from: '../README.md' },
      { from: 'img/', to: 'img/' },
      { from: 'partials/', to: 'partials/' },
    ]),
    new CleanWebpackPlugin({cleanOnceBeforeBuildPatterns: ['**/*', '!cassandra-plugin_*']})
  ],
  resolve: {
    extensions: [".ts", '.tsx', ".js"]
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/, 
        loaders: ["ts-loader"], 
        exclude: /node_modules/,
      },
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader'],
      },
    ]
  }
}
