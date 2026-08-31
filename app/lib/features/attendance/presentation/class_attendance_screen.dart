import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../classes/data/class_repository.dart';
import '../data/attendance_repository.dart';

const _statuses = ['present', 'absent', 'late', 'excused'];

class ClassAttendanceScreen extends ConsumerStatefulWidget {
  const ClassAttendanceScreen({super.key, required this.classId});
  final String classId;

  @override
  ConsumerState<ClassAttendanceScreen> createState() => _ClassAttendanceScreenState();
}

class _ClassAttendanceScreenState extends ConsumerState<ClassAttendanceScreen> {
  final Map<String, String> _statusByStudent = {};
  bool _loaded = false;
  bool _submitting = false;

  @override
  Widget build(BuildContext context) {
    // Fetched via the class-scoped roster endpoint, not the general
    // activity-scoped student list — a substitute coach with no broader
    // activity membership is still authorized for this one class and
    // must be able to see who's enrolled to mark their attendance.
    final rosterAsync = ref.watch(_rosterProvider(widget.classId));
    final existingAsync = ref.watch(_existingMarksProvider(widget.classId));

    return Scaffold(
      appBar: AppBar(title: const Text('Mark attendance')),
      body: rosterAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, __) => Center(child: Text('Failed to load students.\n$e')),
        data: (roster) {
          return existingAsync.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, __) => Center(child: Text('Failed to load attendance.\n$e')),
            data: (existing) {
              if (!_loaded) {
                for (final s in roster) {
                  _statusByStudent[s.id] = 'present';
                }
                for (final e in existing) {
                  _statusByStudent[e.studentId] = e.status;
                }
                _loaded = true;
              }

              if (roster.isEmpty) {
                return const Center(child: Text('No students enrolled in this batch.', style: TextStyle(color: Colors.grey)));
              }

              return Column(
                children: [
                  Expanded(
                    child: ListView.separated(
                      padding: const EdgeInsets.all(16),
                      itemCount: roster.length,
                      separatorBuilder: (_, __) => const SizedBox(height: 8),
                      itemBuilder: (context, index) {
                        final student = roster[index];
                        final current = _statusByStudent[student.id] ?? 'present';
                        return Card(
                          child: Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                            child: Row(
                              children: [
                                Expanded(
                                  child: Text(student.name, style: const TextStyle(fontWeight: FontWeight.w600)),
                                ),
                                DropdownButton<String>(
                                  value: current,
                                  underline: const SizedBox.shrink(),
                                  items: _statuses
                                      .map((s) => DropdownMenuItem(value: s, child: Text(_label(s))))
                                      .toList(),
                                  onChanged: (v) {
                                    if (v == null) return;
                                    setState(() => _statusByStudent[student.id] = v);
                                  },
                                ),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                  SafeArea(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: SizedBox(
                        width: double.infinity,
                        child: ElevatedButton(
                          onPressed: _submitting ? null : () => _submit(roster.map((s) => s.id).toList()),
                          child: _submitting
                              ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                              : const Text('Submit attendance'),
                        ),
                      ),
                    ),
                  ),
                ],
              );
            },
          );
        },
      ),
    );
  }

  String _label(String status) => status[0].toUpperCase() + status.substring(1);

  Future<void> _submit(List<String> studentIds) async {
    setState(() => _submitting = true);
    try {
      final entries = studentIds.map((id) => MarkEntry(studentId: id, status: _statusByStudent[id] ?? 'present')).toList();
      await ref.read(attendanceRepositoryProvider).markBulk(widget.classId, entries);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Attendance submitted.')));
        Navigator.of(context).maybePop();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed to submit attendance: $e')));
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }
}

final _rosterProvider = FutureProvider.autoDispose.family((ref, String classId) {
  return ref.watch(classRepositoryProvider).roster(classId);
});

final _existingMarksProvider = FutureProvider.autoDispose.family((ref, String classId) {
  return ref.watch(attendanceRepositoryProvider).forClass(classId);
});
