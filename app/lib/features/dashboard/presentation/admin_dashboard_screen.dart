import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../../core/widgets/stat_card.dart';
import '../../analytics/data/analytics_repository.dart';

class AdminDashboardScreen extends ConsumerWidget {
  const AdminDashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final summary = ref.watch(overallSummaryProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Admin Dashboard')),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(overallSummaryProvider.future),
        child: AsyncValueWidget(
          value: summary,
          data: (s) => ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text("Today", style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700)),
              const SizedBox(height: 12),
              GridView.count(
                crossAxisCount: 2,
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                mainAxisSpacing: 12,
                crossAxisSpacing: 12,
                childAspectRatio: 1.4,
                children: [
                  StatCard(
                    label: "Today's classes",
                    value: '${s.todayClasses}',
                    icon: Icons.event_note,
                  ),
                  StatCard(
                    label: 'Missing attendance',
                    value: '${s.todayMissingAttendance}',
                    icon: Icons.warning_amber_rounded,
                    color: s.todayMissingAttendance > 0 ? Colors.orange : null,
                  ),
                  StatCard(
                    label: 'Coach check-ins today',
                    value: '${s.coachCheckInsToday}',
                    icon: Icons.location_on_outlined,
                  ),
                  StatCard(
                    label: 'Pending leaves',
                    value: '${s.pendingLeaves}',
                    icon: Icons.event_busy_outlined,
                    onTap: () => context.push('/leaves'),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              Text("This month", style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700)),
              const SizedBox(height: 12),
              GridView.count(
                crossAxisCount: 2,
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                mainAxisSpacing: 12,
                crossAxisSpacing: 12,
                childAspectRatio: 1.4,
                children: [
                  StatCard(
                    label: 'Students',
                    value: '${s.totalStudents}',
                    icon: Icons.groups_outlined,
                  ),
                  StatCard(
                    label: 'Coaches',
                    value: '${s.totalCoaches}',
                    icon: Icons.badge_outlined,
                  ),
                  StatCard(
                    label: 'Attendance rate',
                    value: '${s.attendancePercent.toStringAsFixed(0)}%',
                    icon: Icons.fact_check_outlined,
                    color: Colors.teal,
                  ),
                  StatCard(
                    label: 'Fees collected',
                    value: s.feesCollectedThisMonth.toStringAsFixed(0),
                    icon: Icons.payments_outlined,
                    color: Colors.green,
                  ),
                  StatCard(
                    label: 'Fees pending',
                    value: s.feesPendingThisMonth.toStringAsFixed(0),
                    icon: Icons.pending_actions_outlined,
                    color: Colors.red,
                    onTap: () => context.push('/fees'),
                  ),
                  StatCard(
                    label: 'Activities',
                    value: '${s.totalActivities}',
                    icon: Icons.sports_gymnastics,
                  ),
                ],
              ),
              const SizedBox(height: 24),
              FilledButton.tonalIcon(
                onPressed: () => context.push('/analytics'),
                icon: const Icon(Icons.bar_chart),
                label: const Text('View full analytics'),
              ),
              const SizedBox(height: 80),
            ],
          ),
        ),
      ),
    );
  }
}
