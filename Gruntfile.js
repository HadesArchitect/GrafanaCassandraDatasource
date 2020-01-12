module.exports = function(grunt) {

    grunt.loadNpmTasks('grunt-contrib-clean');
    grunt.loadNpmTasks('grunt-contrib-copy');
    grunt.loadNpmTasks("grunt-ts");

    grunt.initConfig({
  
      pkg: grunt.file.readJSON('package.json'),

      clean: ["dist"],
  
      copy: {
        src_to_dist: {
          cwd: 'src',
          src: ['**', "!**/*.ts"],
          dest: 'dist/',
          expand: true
        },
        readme: {
          expand: true,
          src: ['README.md'],
          dest: 'dist/'
        }
      },

      ts: {
        "build": {
          src: "src/*.ts",
          outDir: "dist/"
        }
      }

    });
  
    grunt.registerTask('default', ['clean', 'copy', 'ts']);
  };
