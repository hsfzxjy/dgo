part of 'ir.dart';

class Param {
  final String name;
  final IR term;

  Param.fromMap(Map m)
      : term = _buildIR(m['Term']),
        name = m['Name'];
}

extension _NullableIRExt on IR? {
  String get dartType => this == null ? 'void' : this!.dartType;
}

extension _IRExt on IR {
  void writeSnippetLoad(GeneratorContext ctx) {
    if (this is Namable && (this as Namable).isNamed) {
      final myUri = (this as Namable).myUri!;
      ctx
        ..sln('{')
        ..sln('$vHolder = ')
        ..sln(ctx.importer.qualifyUri(myUri))
        ..sln('.\$dgoLoad($vArgs);')
        ..sln('}');
    } else {
      writeSnippet$dgoLoad();
    }
  }
}

extension _SumExt<T extends num> on Iterable<T> {
  T sum() => isEmpty ? 0 as T : reduce((a, b) => (a + b) as T);
}

class Method {
  final IR self;
  final String funcName;
  final List<Param> params;
  final int funcId;
  final IR? returnType;
  final bool returnError;

  Method.fromMap(this.self, Map m)
      : funcName = m['Name'],
        funcId = m['FuncId'],
        params = (m['Params'] as List).map((m) => Param.fromMap(m)).toList(),
        returnType = _buildIRNull(m['Return']),
        returnError = m['ReturnError'];

  void writeSnippet(GeneratorContext ctx) {
    var paramSig = params
        .map((p) => '${p.term.outerDartType} ${p.name}')
        .followedBy(['{\$core.Duration? \$timeout,DgoPort? \$port}']).join(',');
    var paramSize = params.map((p) => p.term.goSize).sum();
    paramSize += self.goSize;
    ctx
      ..sln('Future<${returnType.dartType}>')
      ..sln('$funcName($paramSig) async {')
      ..sln('\$port ??= dgo.defaultPort;')
      ..sln('final $vArgs = \$core.List<\$core.dynamic>'
          '.filled($paramSize, null, growable: false);')
      ..sln('var $vIndex = 0;')
      ..sln('$vIndex = \$dgoStore($vArgs, $vIndex);')
      ..for_(
          params,
          (param) => ctx
            ..sln('{')
            ..sln('final $vHolder = ${param.name};')
            ..then(param.term.writeSnippet$dgoStore)
            ..sln('}'))
      ..sln(
        'final \$future = GoMethod($funcId, \$port)'
        '.${returnType == null ? "call" : "callWithResult"}'
        '($vArgs, timeout: \$timeout, hasError: $returnError);',
      )
      ..if_(
        returnType == null,
        () => ctx.sln('return \$future;'),
        else_: () => ctx
          ..sln('{')
          ..sln('final $vArgs = (await \$future).iterator;')
          ..sln('$vArgs.moveNext();')
          ..sln('${returnType!.dartType} $vHolder;')
          ..pipe(returnType!.writeSnippetLoad(ctx))
          ..sln('return $vHolder;')
          ..sln('}'),
      )
      ..sln('}');
  }
}
