module.exports = function(grunt) {

    grunt.loadNpmTasks('grunt-contrib-clean');
    grunt.loadNpmTasks('grunt-contrib-copy');
    grunt.loadNpmTasks("grunt-eslint");

    grunt.initConfig({
  
      pkg: grunt.file.readJSON('package.json'),

      clean: ["dist"],
  
      copy: {
        src_to_dist: {
          cwd: 'src',
          src: '**',
          dest: 'dist/',
          expand: true
        },
        readme: {
          expand: true,
          src: ['README.md'],
          dest: 'dist/'
        }
      },

      eslint: {
        target: 'src/*.ts'
      }
    });
  
    grunt.registerTask('default', ['clean', 'eslint', 'copy']);
};
