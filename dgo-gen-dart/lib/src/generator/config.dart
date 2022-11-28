part of 'generator.dart';

late final Config config;

class Config {
  final JsonMap m;

  const Config(this.m);

  dynamic operator [](String key) => m[key];

  String get projectPath => m['DartProject']['Path'];
  String get generatedInPath =>
      p.join(projectPath, m['DartProject']['GeneratedIn']);
  JsonMap get renames => m['Packages']['Renames'];

  void validate() {
    if (!p.isWithin(projectPath, generatedInPath)) {
      throw 'dgo-gen-dart: generation path $generatedInPath '
          'goes out of project directory $projectPath';
    }
  }
}
