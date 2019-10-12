<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class AlterHistoryTableAddRule extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::table('history', function (Blueprint $table) {
            $table->string('rule', 10)->nullable();
            $table->unsignedInteger('rule_price')->nullable();
            $table->unsignedInteger('current_price')->nullable();
            $table->unsignedInteger('current_price_user_id')->nullable();
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::table('history', function (Blueprint $table) {
            $table->dropColumn('rule');
            $table->dropColumn('rule_price');
            $table->dropColumn('current_price');
            $table->dropColumn('current_price_user_id');
        });
    }
}
