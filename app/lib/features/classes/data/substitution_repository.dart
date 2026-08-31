import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/providers/core_providers.dart';

class SubstitutionModel {
  SubstitutionModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        classId = json['class_id'] as String,
        originalCoachId = json['original_coach_id'] as String,
        substituteCoachId = json['substitute_coach_id'] as String,
        reason = json['reason'] as String? ?? '',
        status = json['status'] as String;

  final String id;
  final String classId;
  final String originalCoachId;
  final String substituteCoachId;
  final String reason;
  final String status;
}

class SubstitutionRepository {
  SubstitutionRepository(this._dio);
  final Dio _dio;

  Future<String> create({required String classId, required String substituteCoachId, required String reason}) async {
    final response = await _dio.post('/substitutions', data: {
      'class_id': classId,
      'substitute_coach_id': substituteCoachId,
      'reason': reason,
    });
    return (response.data as Map<String, dynamic>)['id'] as String;
  }

  Future<void> cancel(String id) => _dio.post('/substitutions/$id/cancel');

  Future<List<SubstitutionModel>> mine() async {
    final response = await _dio.get('/substitutions/mine');
    return (response.data as List<dynamic>).map((e) => SubstitutionModel.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<List<SubstitutionModel>> all() async {
    final response = await _dio.get('/substitutions');
    final data = (response.data as Map<String, dynamic>)['data'] as List<dynamic>;
    return data.map((e) => SubstitutionModel.fromJson(e as Map<String, dynamic>)).toList();
  }
}

final substitutionRepositoryProvider = Provider<SubstitutionRepository>((ref) {
  return SubstitutionRepository(ref.watch(apiClientProvider).dio);
});

final mySubstitutionsProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(substitutionRepositoryProvider).mine();
});

final allSubstitutionsProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(substitutionRepositoryProvider).all();
});
