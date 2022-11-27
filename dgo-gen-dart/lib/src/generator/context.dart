part of 'generator.dart';

@immutable
class GeneratorSymbol {
  final String _value;
  const GeneratorSymbol(this._value);

  @override
  String toString() => _value;

  @override
  int get hashCode => _value.hashCode;

  @override
  bool operator ==(Object other) =>
      other is GeneratorSymbol && _value == other._value;
}

class GeneratorContext {
  final StringSink buffer;
  final Importer importer;
  final _symbols = <GeneratorSymbol, String>{};
  final _usedNames = <String>{};

  GeneratorContext(this.buffer, this.importer);

  String operator [](GeneratorSymbol sym) => _symbols[sym]!;

  String pickUnique(String name) {
    var counter = 1;
    var name2 = name;
    while (true) {
      if (!_usedNames.contains(name2)) {
        _usedNames.add(name2);
        return name2;
      }
      name2 = '$name$counter';
      counter++;
    }
  }

  GeneratorContext withSymbol(GeneratorSymbol sym, String value) =>
      GeneratorContext(buffer, importer)
        .._symbols.addAll(_symbols)
        .._symbols[sym] = value
        .._usedNames.addAll(_usedNames)
        .._usedNames.add(value);

  GeneratorContext withSymbols(Map<GeneratorSymbol, String> syms) =>
      GeneratorContext(buffer, importer)
        .._symbols.addAll(_symbols)
        .._symbols.addAll(syms)
        .._usedNames.addAll(_usedNames.followedBy(syms.values));
}

const vArgs = GeneratorSymbol('#!vArgs');
const vIndex = GeneratorSymbol('#!vIndex');
const vHolder = GeneratorSymbol('#!vHolder');

typedef JsonMap = Map<String, dynamic>;
